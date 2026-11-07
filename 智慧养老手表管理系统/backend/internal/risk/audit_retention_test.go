package risk

import (
	"strconv"
	"testing"
	"time"
)

func TestAuditRetentionAndSnapshotIsolation(t *testing.T) {
	m := New()
	for i := 0; i < 2105; i++ {
		m.RecordDenied("192.0.2.1", "13812345678", strconv.Itoa(i))
	}
	m.mu.Lock()
	snapshot, _ := m.snapshotLocked()
	m.mu.Unlock()
	if len(snapshot.Events) != 2000 || snapshot.Events[0].Detail != "105" || snapshot.Events[1999].Detail != "2104" {
		t.Fatal("audit retention must preserve the latest 2000 events in recording order")
	}
	history, _ := m.History(200)
	if len(history) != 200 || history[0].Detail != "2104" || history[199].Detail != "1905" {
		t.Fatal("history must return the latest events first")
	}
	history[0].Detail = "changed by caller"
	for i := 2105; i < 4105; i++ {
		m.RecordDenied("192.0.2.1", "13812345678", strconv.Itoa(i))
	}
	if snapshot.Events[0].Detail != "105" || snapshot.Events[1999].Detail != "2104" {
		t.Fatal("new events or retention changed an in-flight persistence snapshot")
	}
	history, _ = m.History(1)
	if history[0].Detail != "4104" || history[0].Account != "138****5678" {
		t.Fatal("history changed through an exported slice or lost account masking")
	}
}

func TestAuditExpirySurvivesRestart(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	m := New()
	m.now = func() time.Time { return now }
	if err := m.Persist(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = m.Close() })
	m.RecordDenied("192.0.2.1", "13812345678", "expired")
	now = now.Add(31 * 24 * time.Hour)
	m.RecordDenied("192.0.2.1", "13812345678", "current")
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
	fresh := New()
	fresh.now = func() time.Time { return now }
	if err := fresh.Persist(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := fresh.Close(); err != nil {
			t.Error(err)
		}
	})
	history, _ := fresh.History(200)
	if len(history) != 1 || history[0].Detail != "current" {
		t.Fatal("expired events reappeared after flushing and restarting")
	}
}

func BenchmarkAuditEventAtCapacity(b *testing.B) {
	m := New()
	for i := 0; i < 2000; i++ {
		m.RecordDenied("192.0.2.1", "13812345678", "invalid_credentials")
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.RecordDenied("192.0.2.1", "13812345678", "invalid_credentials")
	}
}
