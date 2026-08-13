package trustclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestCrashStoreStageIsMonotonicAcrossRestart catches an overwriteable
// highest pointer, an equal-version conflict, or an activate that can race a
// newer durable stage and reactivate old bytes.
func TestCrashStoreStageIsMonotonicAcrossRestart(t *testing.T) {
	directory := t.TempDir()
	store, err := NewDirectoryStore(directory)
	if err != nil {
		t.Fatalf("NewDirectoryStore: %v", err)
	}
	ctx := context.Background()
	if err := store.StageAndAdvance(ctx, "audience-a", 2, []byte(`{"version":"2"}`)); err != nil {
		t.Fatalf("stage 2: %v", err)
	}
	for _, test := range []struct {
		version uint64
		body    []byte
	}{
		{version: 1, body: []byte(`{"version":"1"}`)},
		{version: 2, body: []byte(`{"version":"conflict"}`)},
	} {
		if err := store.StageAndAdvance(ctx, "audience-a", test.version, test.body); !errors.Is(err, ErrRollback) {
			t.Fatalf("stage %d error = %v, want rollback", test.version, err)
		}
	}

	restarted, err := NewDirectoryStore(directory)
	if err != nil {
		t.Fatalf("restart: %v", err)
	}
	if highest, err := restarted.LoadHighest(ctx, "audience-a"); err != nil || highest != 2 {
		t.Fatalf("highest after restart = %d, err %v", highest, err)
	}
	if err := restarted.Activate(ctx, "audience-a", 1); !errors.Is(err, ErrRollback) {
		t.Fatalf("activate lower error = %v", err)
	}
	if err := restarted.Activate(ctx, "audience-a", 2); err != nil {
		t.Fatalf("activate 2: %v", err)
	}
	version, body, err := restarted.LoadActive(ctx, "audience-a")
	if err != nil || version != 2 || !bytes.Equal(body, []byte(`{"version":"2"}`)) {
		t.Fatalf("active = %d/%q, err %v", version, body, err)
	}
	if err := restarted.StageAndAdvance(ctx, "audience-a", 3, []byte(`{"version":"3"}`)); err != nil {
		t.Fatalf("stage 3: %v", err)
	}
	if _, _, err := restarted.LoadActive(ctx, "audience-a"); !errors.Is(err, ErrStateNotFound) {
		t.Fatalf("active with pending stage error = %v, want fail-closed state-not-found", err)
	}
	if err := restarted.Activate(ctx, "audience-a", 2); !errors.Is(err, ErrRollback) {
		t.Fatalf("reactivate 2 with stage 3 pending = %v, want rollback", err)
	}
	if err := restarted.Activate(ctx, "audience-a", 3); err != nil {
		t.Fatalf("replace active with 3: %v", err)
	}
	version, body, err = restarted.LoadActive(ctx, "audience-a")
	if err != nil || version != 3 || !bytes.Equal(body, []byte(`{"version":"3"}`)) {
		t.Fatalf("replaced active = %d/%q, err %v", version, body, err)
	}
}

// TestCrashStoreCrossProcessMonotonicity catches relying on an instance-local
// mutex for monotonic staging or activation across separate client processes.
func TestCrashStoreCrossProcessMonotonicity(t *testing.T) {
	directory := t.TempDir()
	stageOne, stageOneOutput := startStoreHelper(t, directory, "stage", 1)
	stageTwo, stageTwoOutput := startStoreHelper(t, directory, "stage", 2)
	if err := stageOne.Wait(); err != nil {
		t.Fatalf("stage 1 helper: %v: %s", err, stageOneOutput.String())
	}
	if err := stageTwo.Wait(); err != nil {
		t.Fatalf("stage 2 helper: %v: %s", err, stageTwoOutput.String())
	}

	activateOne, activateOneOutput := startStoreHelper(t, directory, "activate", 1)
	activateTwo, activateTwoOutput := startStoreHelper(t, directory, "activate", 2)
	if err := activateOne.Wait(); err != nil {
		t.Fatalf("activate 1 helper: %v: %s", err, activateOneOutput.String())
	}
	if err := activateTwo.Wait(); err != nil {
		t.Fatalf("activate 2 helper: %v: %s", err, activateTwoOutput.String())
	}

	store, err := NewDirectoryStore(directory)
	if err != nil {
		t.Fatalf("NewDirectoryStore: %v", err)
	}
	version, body, err := store.LoadActive(context.Background(), "audience-a")
	if err != nil || version != 2 || !bytes.Equal(body, []byte(`{"version":"2"}`)) {
		t.Fatalf("cross-process active = %d/%q, err %v", version, body, err)
	}
}

func TestCrashStoreCrossProcessHelper(t *testing.T) {
	action := os.Getenv("TASK15_STORE_ACTION")
	if action == "" {
		return
	}
	version, err := strconv.ParseUint(os.Getenv("TASK15_STORE_VERSION"), 10, 64)
	if err != nil || (version != 1 && version != 2) {
		t.Fatal("invalid helper version")
	}
	store, err := NewDirectoryStore(os.Getenv("TASK15_STORE_DIRECTORY"))
	if err != nil {
		t.Fatalf("NewDirectoryStore helper: %v", err)
	}
	switch action {
	case "stage":
		err = store.StageAndAdvance(context.Background(), "audience-a", version, []byte(fmt.Sprintf(`{"version":"%d"}`, version)))
		if version == 1 && errors.Is(err, ErrRollback) {
			return
		}
	case "activate":
		err = store.Activate(context.Background(), "audience-a", version)
		if version == 1 && errors.Is(err, ErrRollback) {
			return
		}
	default:
		t.Fatal("invalid helper action")
	}
	if err != nil {
		t.Fatalf("%s %d: %v", action, version, err)
	}
}

func startStoreHelper(t *testing.T, directory, action string, version uint64) (*exec.Cmd, *bytes.Buffer) {
	t.Helper()
	command := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestCrashStoreCrossProcessHelper$") //nolint:gosec // Fixed current test binary.
	command.Env = append(os.Environ(),
		"TASK15_STORE_ACTION="+action,
		"TASK15_STORE_DIRECTORY="+directory,
		"TASK15_STORE_VERSION="+strconv.FormatUint(version, 10),
	)
	output := new(bytes.Buffer)
	command.Stdout = output
	command.Stderr = output
	if err := command.Start(); err != nil {
		t.Fatalf("start %s %d helper: %v", action, version, err)
	}
	return command, output
}

// TestCrashStoreContextAndFiniteBounds catches filesystem work after
// cancellation or unbounded keys/config bytes entering path/allocation logic.
func TestCrashStoreContextAndFiniteBounds(t *testing.T) {
	store, err := NewDirectoryStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewDirectoryStore: %v", err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := store.StageAndAdvance(cancelled, "audience", 1, []byte(`{}`)); !errors.Is(err, ErrStore) {
		t.Fatalf("cancelled stage error = %v", err)
	}
	if err := store.StageAndAdvance(context.Background(), strings.Repeat("x", 257), 1, []byte(`{}`)); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("long key error = %v", err)
	}
	if err := store.StageAndAdvance(context.Background(), "audience", 1, bytes.Repeat([]byte{'x'}, (64<<10)+1)); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("large body error = %v", err)
	}
}

// TestCrashStoreSensitiveSurfaceIsRedacted catches a state directory or active
// plaintext escaping through fmt, slog, JSON, or raw wrapped errors.
func TestCrashStoreSensitiveSurfaceIsRedacted(t *testing.T) {
	const canary = "TASK15-STORE-CANARY"
	store, err := NewDirectoryStore(filepath.Join(t.TempDir(), canary))
	if err != nil {
		t.Fatalf("NewDirectoryStore: %v", err)
	}
	formatted := fmt.Sprintf("%v %#v %+v", store, store, store)
	logged := slog.Any("store", store).Value.Resolve().String()
	if strings.Contains(formatted, canary) || strings.Contains(logged, canary) {
		t.Fatal("store rendered its path")
	}
	if _, err := json.Marshal(store); err == nil {
		t.Fatal("generic JSON accepted store")
	}
}
