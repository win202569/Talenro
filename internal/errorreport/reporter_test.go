package errorreport

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestReportDTOHasOnlyTheFixedScalarAllowlist(t *testing.T) {
	typeOfReport := reflect.TypeOf(Report{})
	want := []struct {
		name   string
		typeOf reflect.Type
	}{
		{name: "Event", typeOf: reflect.TypeOf(Event(""))},
		{name: "Category", typeOf: reflect.TypeOf(Category(""))},
		{name: "Component", typeOf: reflect.TypeOf(Component(""))},
		{name: "Outcome", typeOf: reflect.TypeOf(Outcome(""))},
		{name: "Fingerprint", typeOf: reflect.TypeOf(Fingerprint(""))},
		{name: "BuildVersion", typeOf: reflect.TypeOf("")},
		{name: "TraceID", typeOf: reflect.TypeOf("")},
	}
	if typeOfReport.NumField() != len(want) {
		t.Fatalf("Report field count = %d, want %d", typeOfReport.NumField(), len(want))
	}
	for index, expected := range want {
		field := typeOfReport.Field(index)
		if field.Name != expected.name || field.Type != expected.typeOf {
			t.Fatalf("Report field %d = %s %v, want %s %v", index, field.Name, field.Type, expected.name, expected.typeOf)
		}
		if field.Type.Kind() == reflect.Map || field.Type.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			t.Fatalf("Report field %s accepts unbounded or raw error data", field.Name)
		}
	}
}

func TestReporterBatchesAtMostTwentyAndUsesIndependentTimeouts(t *testing.T) {
	provider := &task18ReportProvider{batches: make(chan []Report, 8), blockFirst: make(chan struct{}), firstStarted: make(chan struct{})}
	observer := newTask18ReportObserver()
	reporter, err := New(t.Context(), provider, 100, 20, 100*time.Millisecond, observer)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { reporter.Close(context.Background()) })

	if !reporter.TryReport(validTask18Report()) {
		t.Fatal("first valid report was rejected")
	}
	select {
	case <-provider.firstStarted:
	case <-time.After(time.Second):
		t.Fatal("report provider did not start")
	}
	for range 25 {
		if !reporter.TryReport(validTask18Report()) {
			t.Fatal("report within queue capacity was rejected")
		}
	}
	close(provider.blockFirst)

	total := 0
	deadline := time.After(2 * time.Second)
	for total < 26 {
		select {
		case batch := <-provider.batches:
			if len(batch) == 0 || len(batch) > 20 {
				t.Fatalf("provider batch size = %d, want 1..20", len(batch))
			}
			total += len(batch)
		case <-deadline:
			t.Fatalf("provider received %d reports, want 26", total)
		}
	}
	if observer.count(ResultSent) == 0 {
		t.Fatal("successful provider delivery was not counted")
	}
}

func TestReporterTryReportIsNonblockingAndDropsWhenQueueIsFull(t *testing.T) {
	provider := &task18ReportProvider{blockFirst: make(chan struct{}), firstStarted: make(chan struct{})}
	observer := newTask18ReportObserver()
	reporter, err := New(t.Context(), provider, 10, 10, time.Second, observer)
	if err != nil {
		t.Fatal(err)
	}
	if !reporter.TryReport(validTask18Report()) {
		t.Fatal("first valid report was rejected")
	}
	select {
	case <-provider.firstStarted:
	case <-time.After(time.Second):
		t.Fatal("provider did not block on the first report")
	}
	for range 10 {
		if !reporter.TryReport(validTask18Report()) {
			t.Fatal("queue rejected a report before reaching capacity")
		}
	}
	started := time.Now()
	if reporter.TryReport(validTask18Report()) {
		t.Fatal("full queue accepted another report")
	}
	if elapsed := time.Since(started); elapsed > 100*time.Millisecond {
		t.Fatalf("full-queue report blocked for %s", elapsed)
	}
	if observer.count(ResultDropped) != 1 {
		t.Fatalf("drop count = %d, want 1", observer.count(ResultDropped))
	}
	close(provider.blockFirst)
	reporter.Close(context.Background())
}

func TestReporterContainsProviderPanicAndRejectsInvalidOrClosedReports(t *testing.T) {
	provider := &task18ReportProvider{panicFirst: true, batches: make(chan []Report, 2)}
	observer := newTask18ReportObserver()
	reporter, err := New(t.Context(), provider, 10, 10, 100*time.Millisecond, observer)
	if err != nil {
		t.Fatal(err)
	}
	if reporter.TryReport(Report{Event: Event("CANARY"), TraceID: "CANARY"}) {
		t.Fatal("invalid report entered the queue")
	}
	if !reporter.TryReport(validTask18Report()) {
		t.Fatal("valid report was rejected")
	}
	deadline := time.Now().Add(time.Second)
	for observer.count(ResultProviderFailed) == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if observer.count(ResultProviderFailed) != 1 {
		t.Fatal("provider panic was not contained and counted")
	}
	if !reporter.TryReport(validTask18Report()) {
		t.Fatal("worker stopped after provider panic")
	}
	select {
	case <-provider.batches:
	case <-time.After(time.Second):
		t.Fatal("worker did not continue after provider panic")
	}
	reporter.Close(context.Background())
	if reporter.TryReport(validTask18Report()) {
		t.Fatal("closed reporter accepted a report")
	}
}

func TestReporterCloseHonorsShutdownDeadlineAndDropsRemainingQueue(t *testing.T) {
	provider := &task18ReportProvider{waitForContext: true, firstStarted: make(chan struct{})}
	observer := newTask18ReportObserver()
	reporter, err := New(t.Context(), provider, 10, 10, time.Second, observer)
	if err != nil {
		t.Fatal(err)
	}
	if !reporter.TryReport(validTask18Report()) {
		t.Fatal("first valid report was rejected")
	}
	select {
	case <-provider.firstStarted:
	case <-time.After(time.Second):
		t.Fatal("provider did not start")
	}
	for range 5 {
		if !reporter.TryReport(validTask18Report()) {
			t.Fatal("queued report was rejected")
		}
	}
	shutdown, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	started := time.Now()
	reporter.Close(shutdown)
	if elapsed := time.Since(started); elapsed > 250*time.Millisecond {
		t.Fatalf("reporter close exceeded shutdown budget: %s", elapsed)
	}
	if reporter.TryReport(validTask18Report()) {
		t.Fatal("closing reporter accepted new work")
	}
	if observer.count(ResultDropped) < 5 {
		t.Fatalf("shutdown dropped %d reports, want at least 5", observer.count(ResultDropped))
	}
}

func validTask18Report() Report {
	return Report{
		Event: EventRequestFailure, Category: CategoryDependency,
		Component: ComponentControlAPI, Outcome: OutcomeFailure,
		Fingerprint: FingerprintDependency, BuildVersion: "v1.2.3",
		TraceID: "0123456789abcdef0123456789abcdef",
	}
}

type task18ReportProvider struct {
	mu             sync.Mutex
	calls          int
	batches        chan []Report
	blockFirst     chan struct{}
	firstStarted   chan struct{}
	panicFirst     bool
	waitForContext bool
}

func (provider *task18ReportProvider) Send(ctx context.Context, reports []Report) error {
	provider.mu.Lock()
	provider.calls++
	call := provider.calls
	provider.mu.Unlock()
	if call == 1 && provider.firstStarted != nil {
		close(provider.firstStarted)
	}
	if call == 1 && provider.panicFirst {
		panic("CANARY provider panic")
	}
	if call == 1 && provider.blockFirst != nil {
		select {
		case <-provider.blockFirst:
		case <-ctx.Done():
			return errors.New("CANARY provider timeout")
		}
	}
	if provider.waitForContext {
		<-ctx.Done()
		return errors.New("CANARY provider canceled")
	}
	if provider.batches != nil {
		provider.batches <- append([]Report(nil), reports...)
	}
	return nil
}

type task18ReportObserver struct {
	mu     sync.Mutex
	counts map[DeliveryResult]int
}

func newTask18ReportObserver() *task18ReportObserver {
	return &task18ReportObserver{counts: make(map[DeliveryResult]int)}
}

func (observer *task18ReportObserver) Record(_ Component, result DeliveryResult) {
	observer.mu.Lock()
	defer observer.mu.Unlock()
	observer.counts[result]++
}

func (observer *task18ReportObserver) count(result DeliveryResult) int {
	observer.mu.Lock()
	defer observer.mu.Unlock()
	return observer.counts[result]
}
