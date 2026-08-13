package trustclient

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"
)

const (
	maximumStoredPayloadBytes = 64 << 10
	maximumStateEntries       = 4096
	stateKeyDomain            = "TALENRO-CLIENT-STATE-KEY-V1\x00"
	stagePrefix               = "stage-"
	activePrefix              = "active-"
	stateSuffix               = ".bin"
	storeMagic                = "TLSTG001"
	storeHeaderBytes          = 8 + 8 + 4 + 32
)

// DirectoryStore durably retains immutable staged versions and one atomic
// active snapshot below a caller-owned directory.
type DirectoryStore struct {
	root string
	mu   sync.Mutex
}

// NewDirectoryStore creates a bounded durable client state directory.
func NewDirectoryStore(root string) (*DirectoryStore, error) {
	if len(root) == 0 || len(root) > 4096 || !utf8.ValidString(root) {
		return nil, ErrInvalidArgument
	}
	absolute, err := filepath.Abs(root)
	if err != nil || absolute == "" {
		return nil, ErrInvalidArgument
	}
	if err := os.MkdirAll(absolute, 0o700); err != nil {
		return nil, ErrStore
	}
	return &DirectoryStore{root: absolute}, nil
}

// LoadHighest returns the highest immutable staged version, or zero when the
// bounded state key has never been staged.
func (store *DirectoryStore) LoadHighest(ctx context.Context, key string) (uint64, error) {
	if store == nil || ctx == nil || ctx.Err() != nil || !validStateKey(key) {
		return 0, storeArgumentError(ctx, key)
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	return loadHighestDirectory(store.stateDirectory(key))
}

// StageAndAdvance fsyncs an immutable version file and its directory. Existing
// equal or higher versions are never replaced.
func (store *DirectoryStore) StageAndAdvance(ctx context.Context, key string, version uint64, body []byte) error {
	if store == nil || ctx == nil || ctx.Err() != nil || !validStateKey(key) || version == 0 ||
		len(body) < 2 || len(body) > maximumStoredPayloadBytes {
		return storeArgumentError(ctx, key)
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	directory := store.stateDirectory(key)
	if err := os.MkdirAll(directory, 0o700); err != nil { //nolint:gosec // Directory is below the validated private root and named by a SHA-256 state-key digest.
		return ErrStore
	}
	if err := syncDirectory(store.root); err != nil {
		return ErrStore
	}
	highest, err := loadHighestDirectory(directory)
	if err != nil {
		return err
	}
	if version <= highest {
		return ErrRollback
	}
	frame := encodeStoreFrame(version, body)
	defer clear(frame)
	temporary, err := os.CreateTemp(directory, ".stage-*")
	if err != nil {
		return ErrStore
	}
	temporaryName := temporary.Name()
	completed := false
	defer func() {
		_ = temporary.Close()
		if !completed {
			_ = os.Remove(temporaryName) //nolint:gosec // CreateTemp returned this path inside the private state directory.
		}
	}()
	if err := temporary.Chmod(0o600); err != nil || writeAll(temporary, frame) != nil || temporary.Sync() != nil || temporary.Close() != nil {
		return ErrStore
	}
	if ctx.Err() != nil {
		return ErrStore
	}
	target := filepath.Join(directory, stageFilename(version))
	if _, err := os.Stat(target); err == nil { //nolint:gosec // Target is below the private digest-named directory with a numeric filename.
		return ErrRollback
	} else if !os.IsNotExist(err) {
		return ErrStore
	}
	if err := os.Link(temporaryName, target); err != nil {
		if _, statErr := os.Stat(target); statErr == nil { //nolint:gosec // Target is below the private digest-named directory with a numeric filename.
			return ErrRollback
		}
		return ErrStore
	}
	if err := os.Remove(temporaryName); err != nil { //nolint:gosec // CreateTemp returned this path inside the private state directory.
		return ErrStore
	}
	completed = true
	if err := syncDirectory(directory); err != nil {
		return ErrStore
	}
	committed, err := loadHighestDirectory(directory)
	if err != nil {
		return err
	}
	if committed > version {
		return ErrRollback
	}
	if committed != version {
		return ErrStore
	}
	return nil
}

// Activate atomically publishes an immutable activation marker only for the
// current highest staged version. LoadActive resolves the highest marker and
// fails closed while a newer staged version is pending, so separate processes
// cannot reactivate or observe an older snapshot.
func (store *DirectoryStore) Activate(ctx context.Context, key string, version uint64) error {
	if store == nil || ctx == nil || ctx.Err() != nil || !validStateKey(key) || version == 0 {
		return storeArgumentError(ctx, key)
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	directory := store.stateDirectory(key)
	highest, err := loadHighestDirectory(directory)
	if err != nil {
		return err
	}
	if version != highest {
		return ErrRollback
	}
	frame, err := readStoreFrame(filepath.Join(directory, stageFilename(version)))
	if err != nil {
		clear(frame)
		return ErrStore
	}
	decodedVersion, body, ok := decodeStoreFrame(frame)
	if !ok || decodedVersion != version {
		clear(frame)
		clear(body)
		return ErrStore
	}
	clear(body)
	clear(frame)
	if ctx.Err() != nil {
		return ErrStore
	}
	marker := filepath.Join(directory, activeFilename(version))
	if err := os.Link(filepath.Join(directory, stageFilename(version)), marker); err != nil {
		if _, statErr := os.Stat(marker); statErr != nil { //nolint:gosec // Marker is below the private digest-named directory with a numeric filename.
			return ErrStore
		}
		markerFrame, readErr := readStoreFrame(marker)
		markerVersion, markerBody, valid := decodeStoreFrame(markerFrame)
		clear(markerFrame)
		clear(markerBody)
		if readErr != nil || !valid || markerVersion != version {
			return ErrStore
		}
	}
	if err := syncDirectory(directory); err != nil {
		return ErrStore
	}
	committed, err := loadHighestDirectory(directory)
	if err != nil {
		return err
	}
	if committed > version {
		return ErrRollback
	}
	if committed != version {
		return ErrStore
	}
	return nil
}

// LoadActive reads and authenticates the current active snapshot.
func (store *DirectoryStore) LoadActive(ctx context.Context, key string) (uint64, []byte, error) {
	if store == nil || ctx == nil || ctx.Err() != nil || !validStateKey(key) {
		return 0, nil, storeArgumentError(ctx, key)
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	directory := store.stateDirectory(key)
	highest, err := loadHighestDirectory(directory)
	if err != nil {
		return 0, nil, err
	}
	active, err := loadHighestVersion(directory, activePrefix)
	if err != nil {
		return 0, nil, err
	}
	if active == 0 || active != highest {
		return 0, nil, ErrStateNotFound
	}
	frame, err := readStoreFrame(filepath.Join(directory, activeFilename(active)))
	if err != nil {
		clear(frame)
		return 0, nil, ErrStore
	}
	version, body, ok := decodeStoreFrame(frame)
	clear(frame)
	if !ok {
		clear(body)
		return 0, nil, ErrStore
	}
	return version, body, nil
}

func (store *DirectoryStore) stateDirectory(key string) string {
	material := append([]byte(stateKeyDomain), key...)
	digest := sha256.Sum256(material)
	clear(material)
	return filepath.Join(store.root, hex.EncodeToString(digest[:]))
}

// Format redacts every fmt rendering of durable client state.
func (*DirectoryStore) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("trustclient.DirectoryStore([REDACTED])"))
}

// LogValue redacts structured logging of durable client state.
func (*DirectoryStore) LogValue() slog.Value {
	return slog.StringValue("trustclient.DirectoryStore([REDACTED])")
}

// MarshalJSON rejects generic serialization of durable client state.
func (*DirectoryStore) MarshalJSON() ([]byte, error) { return nil, ErrInvalidArgument }

// UnmarshalJSON rejects untrusted construction of durable client state.
func (*DirectoryStore) UnmarshalJSON([]byte) error { return ErrInvalidArgument }

type memoryEntry struct {
	version uint64
	body    []byte
}

// MemoryStore is a finite concurrency-safe store for tests and ephemeral
// conformance callers. It enforces the same monotonic ordering as DirectoryStore.
type MemoryStore struct {
	mu     sync.Mutex
	staged map[string]memoryEntry
	active map[string]memoryEntry
}

// NewMemoryStore constructs an empty bounded in-memory Store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{staged: make(map[string]memoryEntry), active: make(map[string]memoryEntry)}
}

// LoadHighest implements Store.
func (store *MemoryStore) LoadHighest(ctx context.Context, key string) (uint64, error) {
	if store == nil || ctx == nil || ctx.Err() != nil || !validStateKey(key) {
		return 0, storeArgumentError(ctx, key)
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	return store.staged[key].version, nil
}

// StageAndAdvance implements Store.
func (store *MemoryStore) StageAndAdvance(ctx context.Context, key string, version uint64, body []byte) error {
	if store == nil || ctx == nil || ctx.Err() != nil || !validStateKey(key) || version == 0 || len(body) < 2 || len(body) > maximumStoredPayloadBytes {
		return storeArgumentError(ctx, key)
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if version <= store.staged[key].version {
		return ErrRollback
	}
	store.staged[key] = memoryEntry{version: version, body: append([]byte(nil), body...)}
	return nil
}

// Activate implements Store.
func (store *MemoryStore) Activate(ctx context.Context, key string, version uint64) error {
	if store == nil || ctx == nil || ctx.Err() != nil || !validStateKey(key) || version == 0 {
		return storeArgumentError(ctx, key)
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	staged := store.staged[key]
	if version != staged.version {
		return ErrRollback
	}
	store.active[key] = memoryEntry{version: staged.version, body: append([]byte(nil), staged.body...)}
	return nil
}

// Format redacts every fmt rendering of in-memory client state.
func (*MemoryStore) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("trustclient.MemoryStore([REDACTED])"))
}

// LogValue redacts structured logging of in-memory client state.
func (*MemoryStore) LogValue() slog.Value {
	return slog.StringValue("trustclient.MemoryStore([REDACTED])")
}

// MarshalJSON rejects generic serialization of in-memory client state.
func (*MemoryStore) MarshalJSON() ([]byte, error) { return nil, ErrInvalidArgument }

// UnmarshalJSON rejects untrusted construction of in-memory client state.
func (*MemoryStore) UnmarshalJSON([]byte) error { return ErrInvalidArgument }

func loadHighestDirectory(directory string) (uint64, error) {
	highest, err := loadHighestVersion(directory, stagePrefix)
	if err != nil || highest == 0 {
		return highest, err
	}
	frame, err := readStoreFrame(filepath.Join(directory, stageFilename(highest)))
	if err != nil {
		clear(frame)
		return 0, ErrStore
	}
	version, body, ok := decodeStoreFrame(frame)
	clear(frame)
	clear(body)
	if !ok || version != highest {
		return 0, ErrStore
	}
	return highest, nil
}

func loadHighestVersion(directory, prefix string) (uint64, error) {
	entries, err := readBoundedDirectory(directory)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, ErrStore
	}
	var highest uint64
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, stateSuffix) {
			continue
		}
		number := strings.TrimSuffix(strings.TrimPrefix(name, prefix), stateSuffix)
		if len(number) != 20 {
			return 0, ErrStore
		}
		version, parseErr := strconv.ParseUint(number, 10, 64)
		if parseErr != nil || version == 0 {
			return 0, ErrStore
		}
		if version > highest {
			highest = version
		}
	}
	return highest, nil
}

func stageFilename(version uint64) string {
	return fmt.Sprintf("%s%020d%s", stagePrefix, version, stateSuffix)
}

func activeFilename(version uint64) string {
	return fmt.Sprintf("%s%020d%s", activePrefix, version, stateSuffix)
}

func encodeStoreFrame(version uint64, body []byte) []byte {
	frame := make([]byte, storeHeaderBytes+len(body))
	copy(frame[:8], storeMagic)
	binary.BigEndian.PutUint64(frame[8:16], version)
	binary.BigEndian.PutUint32(frame[16:20], uint32(len(body))) //nolint:gosec // Stored payloads are bounded to 64 KiB.
	digest := sha256.Sum256(body)
	copy(frame[20:52], digest[:])
	copy(frame[storeHeaderBytes:], body)
	return frame
}

func decodeStoreFrame(frame []byte) (uint64, []byte, bool) {
	if len(frame) < storeHeaderBytes || string(frame[:8]) != storeMagic {
		return 0, nil, false
	}
	version := binary.BigEndian.Uint64(frame[8:16])
	length := int(binary.BigEndian.Uint32(frame[16:20]))
	if version == 0 || length < 2 || length > maximumStoredPayloadBytes || length != len(frame)-storeHeaderBytes {
		return 0, nil, false
	}
	body := append([]byte(nil), frame[storeHeaderBytes:]...)
	digest := sha256.Sum256(body)
	if !bytes.Equal(frame[20:52], digest[:]) {
		clear(body)
		return 0, nil, false
	}
	return version, body, true
}

func writeAll(writer io.Writer, body []byte) error {
	for len(body) > 0 {
		written, err := writer.Write(body)
		if err != nil || written <= 0 || written > len(body) {
			return ErrStore
		}
		body = body[written:]
	}
	return nil
}

func readStoreFrame(path string) ([]byte, error) {
	file, err := os.Open(path) //nolint:gosec // The path is derived from the private store root and hashed state key.
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	information, err := file.Stat()
	maximum := int64(storeHeaderBytes + maximumStoredPayloadBytes)
	if err != nil || !information.Mode().IsRegular() || information.Size() < storeHeaderBytes || information.Size() > maximum {
		return nil, ErrStore
	}
	frame, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil || int64(len(frame)) > maximum {
		clear(frame)
		return nil, ErrStore
	}
	return frame, nil
}

func readBoundedDirectory(directory string) ([]os.DirEntry, error) {
	handle, err := os.Open(directory) //nolint:gosec // The directory is derived from the private store root and hashed state key.
	if err != nil {
		return nil, err
	}
	defer func() { _ = handle.Close() }()
	entries, err := handle.ReadDir(maximumStateEntries + 1)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	if len(entries) > maximumStateEntries {
		return nil, ErrStore
	}
	return entries, nil
}

func syncDirectory(directory string) error {
	handle, err := os.Open(directory) //nolint:gosec // The directory is the validated private store root or state directory.
	if err != nil {
		return err
	}
	defer func() { _ = handle.Close() }()
	// Some supported platforms do not implement directory fsync. The file and
	// rename are still durable to the extent the platform exposes; where Sync
	// is implemented it is required to succeed.
	if err := handle.Sync(); err != nil && runtime.GOOS != "windows" {
		return err
	}
	return nil
}

func validStateKey(key string) bool {
	return len(key) > 0 && len(key) <= maximumStateKeyBytes && utf8.ValidString(key)
}

func storeArgumentError(ctx context.Context, key string) error {
	if ctx == nil || ctx.Err() != nil {
		return ErrStore
	}
	if !validStateKey(key) {
		return ErrInvalidArgument
	}
	return ErrInvalidArgument
}

var _ Store = (*DirectoryStore)(nil)
var _ Store = (*MemoryStore)(nil)
var _ json.Marshaler = (*DirectoryStore)(nil)
var _ json.Marshaler = (*MemoryStore)(nil)
