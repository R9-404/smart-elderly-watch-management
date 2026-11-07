package risk

import (
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestAuditLifecycleAndConcurrentClose(t *testing.T) {
	dir := t.TempDir()
	m := New()
	if err := m.Persist(dir); err != nil {
		t.Fatal(err)
	}
	if err := m.Persist(dir); err == nil {
		t.Fatal("duplicate initialization accepted")
	}
	var writers sync.WaitGroup
	for i := 0; i < 100; i++ {
		writers.Go(func() { m.FailedLogin("192.0.2.1", "13812345678") })
	}
	writers.Wait()
	for i := 0; i < 8; i++ {
		writers.Go(func() {
			if err := m.Close(); err != nil {
				t.Error(err)
			}
		})
	}
	writers.Wait()
	fresh := New()
	if err := fresh.Persist(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := fresh.Close(); err != nil {
			t.Error(err)
		}
	})
	events, _ := fresh.History(200)
	if len(events) != 100 {
		t.Fatalf("lost shutdown events: %d", len(events))
	}
	if _, err := m.AddRule("192.0.2.1", "block", "confirmed abuse", "admin", time.Hour); err == nil {
		t.Fatal("closed manager accepted rule")
	}
}

func TestAuditRecordingDoesNotWaitForDiskWriter(t *testing.T) {
	m := New()
	if err := m.Persist(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := m.Close(); err != nil {
			t.Error(err)
		}
	})
	m.diskMu.Lock()
	done := make(chan struct{})
	go func() { m.FailedLogin("192.0.2.2", "13812345678"); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		m.diskMu.Unlock()
		t.Fatal("authentication event waited for disk")
	}
	m.diskMu.Unlock()
}

func TestFailedRuleWriteRestoresOrderAndDoesNotReappear(t *testing.T) {
	dir := t.TempDir()
	m := New()
	if err := m.Persist(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := m.Close(); err != nil {
			t.Error(err)
		}
	})
	for _, ip := range []string{"192.0.2.1", "192.0.2.2", "192.0.2.3"} {
		if _, err := m.AddRule(ip, "block", "confirmed abuse", "admin", time.Hour); err != nil {
			t.Fatal(err)
		}
	}
	if err := m.Flush(); err != nil {
		t.Fatal(err)
	}
	_, before := m.History(200)
	// Force replacement to fail on both Windows and Unix, without permissions tricks.
	if err := os.Remove(m.storage); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(m.storage, 0700); err != nil {
		t.Fatal(err)
	}
	if ok, err := m.DeleteRule(before[1].ID, "admin"); err == nil || ok {
		t.Fatal("failed deletion reported success")
	}
	if _, err := m.AddRule("203.0.113.1", "block", "confirmed abuse", "admin", time.Hour); err == nil {
		t.Fatal("failed addition reported success")
	}
	_, after := m.History(200)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("failed write changed in-memory rules/order")
	}
	if err := os.Remove(m.storage); err != nil {
		t.Fatal(err)
	}
	if err := m.Flush(); err != nil {
		t.Fatal(err)
	}
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
	fresh := New()
	if err := fresh.Persist(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := fresh.Close(); err != nil {
			t.Error(err)
		}
	})
	_, restored := fresh.History(200)
	if !reflect.DeepEqual(before, restored) {
		t.Fatal("failed mutation persisted after retry")
	}
	if files, err := filepath.Glob(filepath.Join(dir, ".risk-*.tmp")); err != nil || len(files) != 0 {
		t.Fatalf("temporary files leaked: %v %v", files, err)
	}
}

func TestOlderAuditSnapshotCannotOverwriteNewerRule(t *testing.T) {
	dir := t.TempDir()
	m := New()
	if err := m.Persist(dir); err != nil {
		t.Fatal(err)
	}
	m.FailedLogin("192.0.2.1", "13812345678")
	m.mu.Lock()
	old, generation := m.snapshotLocked()
	m.mu.Unlock()
	if _, err := m.AddRule("203.0.113.1", "block", "confirmed abuse", "admin", time.Hour); err != nil {
		t.Fatal(err)
	}
	if err := m.writeSnapshot(old, generation); err != nil {
		t.Fatal(err)
	}
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
	fresh := New()
	if err := fresh.Persist(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := fresh.Close(); err != nil {
			t.Error(err)
		}
	})
	if fresh.BeforeLogin("203.0.113.1", "a").Allowed() {
		t.Fatal("old snapshot erased persisted rule")
	}
}
