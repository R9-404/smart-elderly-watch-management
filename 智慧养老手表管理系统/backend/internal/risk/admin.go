package risk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"log"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Rule is a managed IP/CIDR block. It is separate from temporary automatic cool-downs.
type Rule struct {
	ID        string    `json:"id"`
	CIDR      string    `json:"cidr"`
	Action    string    `json:"action"` // block; skip-auto
	Reason    string    `json:"reason"`
	Actor     string    `json:"actor"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt,omitempty"`
}

type Event struct {
	At                time.Time  `json:"at"`
	IP                string     `json:"ip"`
	Account           string     `json:"account"` // masked phone; no passwords or tokens
	Type              string     `json:"type"`
	Detail            string     `json:"detail"`
	Device            DeviceInfo `json:"device"`
	Network           string     `json:"network"`
	NetworkConfidence int        `json:"networkConfidence,omitempty"`
	NetworkProvider   string     `json:"networkProvider,omitempty"`
	NetworkSource     string     `json:"networkSource,omitempty"`
}

func masked(account string) string {
	if len(account) < 7 {
		return "***"
	}
	return account[:3] + "****" + account[len(account)-4:]
}

func normalizedPrefix(s string) (netip.Prefix, error) {
	if a, err := netip.ParseAddr(s); err == nil {
		return netip.PrefixFrom(a.Unmap(), a.Unmap().BitLen()), nil
	}
	p, err := netip.ParsePrefix(s)
	if err != nil {
		return netip.Prefix{}, errors.New("invalid IP or CIDR")
	}
	return p.Masked(), nil
}

// Persist uses a private JSON snapshot, shared by restarts of a single process.
// Use a shared database instead if multiple server replicas will be deployed.
// Persist can be initialized only once. Call Close after HTTP requests drain.
func (m *Manager) Persist(dir string) error {
	m.closeMu.Lock()
	defer m.closeMu.Unlock()
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.initialized || m.closed {
		return errors.New("risk manager already initialized or closed")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	storage := filepath.Join(dir, "risk-security.json")
	var v auditSnapshot
	f, err := os.Open(storage)
	if err == nil {
		b, readErr := io.ReadAll(io.LimitReader(f, 16*1024*1024+1))
		closeErr := f.Close()
		if readErr != nil {
			return readErr
		}
		if closeErr != nil {
			return closeErr
		}
		if len(b) > 16*1024*1024 {
			return errors.New("risk snapshot too large")
		}
		if err = json.Unmarshal(b, &v); err != nil {
			return fmt.Errorf("load risk history: %w", err)
		}
		if len(v.Rules) > 4096 {
			return errors.New("too many risk rules")
		}
		for _, rule := range v.Rules {
			if _, err := normalizedPrefix(rule.CIDR); err != nil {
				return err
			}
			if rule.Action != "block" && rule.Action != "skip-auto" {
				return errors.New("invalid stored risk action")
			}
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	m.storage, m.rules, m.events = storage, v.Rules, v.Events
	m.initialized = true
	m.trimEventsLocked()
	ctx, cancel := context.WithCancel(context.Background())
	m.auditWake = make(chan struct{}, 1)
	m.auditDone = make(chan struct{})
	m.auditCancel = cancel
	go m.auditWorker(ctx)
	return nil
}

type auditSnapshot struct {
	Rules  []Rule  `json:"rules"`
	Events []Event `json:"events"`
}

// Called while m.mu is held; it copies mutable slices before releasing the lock.
func (m *Manager) snapshotLocked() (auditSnapshot, uint64) {
	return auditSnapshot{Rules: append([]Rule(nil), m.rules...), Events: append([]Event(nil), m.events...)}, m.generation
}

func (m *Manager) writeSnapshot(s auditSnapshot, gen uint64) error {
	if m.storage == "" {
		return nil
	}
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	m.diskMu.Lock()
	defer m.diskMu.Unlock()
	if gen <= m.savedGeneration {
		return nil
	}
	// The temporary file is unique to this writer and permissions are private.
	f, err := os.CreateTemp(filepath.Dir(m.storage), ".risk-*.tmp")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if _, err = f.Write(b); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(tmp, m.storage); err != nil {
		return err
	}
	m.savedGeneration = gen
	return nil
}

// saveLocked is reserved for low-frequency administrator rule mutations.
// Authentication attempts never call it.
func (m *Manager) saveLocked() error {
	s, gen := m.snapshotLocked()
	return m.writeSnapshot(s, gen)
}

func (m *Manager) notifyAuditLocked() {
	if m.auditWake != nil {
		select {
		case m.auditWake <- struct{}{}:
		default:
		}
	}
}

// Flush forces pending events to disk, for shutdown or integration tests.
func (m *Manager) Flush() error {
	m.mu.Lock()
	snapshot, gen := m.snapshotLocked()
	m.mu.Unlock()
	return m.writeSnapshot(snapshot, gen)
}

// Close flushes pending audit events and stops the writer.
// Call at server shutdown after requests have drained.
func (m *Manager) Close() error {
	m.closeMu.Lock()
	defer m.closeMu.Unlock()
	m.mu.Lock()
	m.closed = true
	cancel, done := m.auditCancel, m.auditDone
	m.auditCancel = nil
	m.mu.Unlock()
	if cancel != nil {
		cancel()
		<-done
	}
	return m.Flush()
}

func (m *Manager) auditWorker(ctx context.Context) {
	defer close(m.auditDone)
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	dirty := false
	for {
		select {
		case <-m.auditWake:
			dirty = true
		case <-ticker.C:
			if dirty {
				if err := m.Flush(); err != nil {
					log.Printf("risk audit persistence error: %v", err)
				} else {
					dirty = false
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (m *Manager) recordEventLocked(ip, account, typ, detail string, signals ...string) {
	if m.closed {
		return
	}
	// Never persist passwords, cookies, headers, or raw phone numbers.
	e := Event{At: m.now().UTC(), IP: ip, Account: masked(account), Type: typ, Detail: detail, Device: IdentifyDevice(signals...)}
	classification := m.classifyNetworkLocked(ip)
	e.Network, e.NetworkConfidence, e.NetworkProvider, e.NetworkSource = classification.Kind, classification.Confidence, classification.Provider, classification.Source
	m.events = append(m.events, e)
	m.trimEventsLocked()
	m.generation++
	m.notifyAuditLocked()
}

func (m *Manager) trimEventsLocked() {
	// Last 2000 entries, maximum 30 days. This limits disk use and retention.
	cutoff := m.now().Add(-30 * 24 * time.Hour)
	first := 0
	for first < len(m.events) && m.events[first].At.Before(cutoff) {
		first++
	}
	if len(m.events)-first > 2000 {
		first = len(m.events) - 2000
	}
	// Drop references to expired records without copying the whole live buffer.
	// History and persistence snapshots already own independent copies.
	clear(m.events[:first])
	m.events = m.events[first:]
}

func (m *Manager) RecordDenied(ip, account, detail string, signals ...string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordEventLocked(ip, account, "request_blocked", detail, signals...)
}

func (m *Manager) matchRuleLocked(ip string) (Decision, bool) {
	a, err := netip.ParseAddr(ip)
	if err != nil {
		return Decision{}, false
	}
	a = a.Unmap()
	now := m.now()
	// Block takes precedence over skip-auto; explicit exceptions only bypass auto throttling.
	skip := false
	for _, rule := range m.rules {
		if !rule.ExpiresAt.IsZero() && !now.Before(rule.ExpiresAt) {
			continue
		}
		p, err := normalizedPrefix(rule.CIDR)
		if err != nil || !p.Contains(a) {
			continue
		}
		if rule.Action == "block" {
			retry := 3600
			if !rule.ExpiresAt.IsZero() {
				retry = int(rule.ExpiresAt.Sub(now).Seconds() + 1)
			}
			return Decision{RetryAfter: retry, RiskLevel: "blocked"}, true
		}
		if rule.Action == "skip-auto" {
			skip = true
		}
	}
	if skip {
		return Decision{RiskLevel: "low"}, true
	}
	return Decision{}, false
}

func (m *Manager) AddRule(raw, action, reason, actor string, ttl time.Duration) (Rule, error) {
	prefix, err := normalizedPrefix(strings.TrimSpace(raw))
	if err != nil {
		return Rule{}, err
	}
	if action != "block" && action != "skip-auto" {
		return Rule{}, errors.New("invalid action")
	}
	if ttl < 0 || ttl > 365*24*time.Hour {
		return Rule{}, errors.New("invalid duration")
	}
	reason = strings.TrimSpace(reason)
	if len(reason) < 3 || len(reason) > 240 {
		return Rule{}, errors.New("reason length must be 3..240")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	now := m.now().UTC()
	if m.closed {
		return Rule{}, errors.New("risk manager closed")
	}
	if len(m.rules) >= 4096 {
		return Rule{}, errors.New("too many risk rules")
	}
	rule := Rule{ID: uuid.NewString(), CIDR: prefix.String(), Action: action, Reason: reason, Actor: actor, CreatedAt: now}
	if ttl > 0 {
		rule.ExpiresAt = now.Add(ttl)
	}
	m.rules = append(m.rules, rule)
	m.generation++
	if err = m.saveLocked(); err != nil {
		m.rules = m.rules[:len(m.rules)-1]
		return Rule{}, err
	}
	m.recordEventLocked(prefix.String(), "", "rule_created", action+": "+reason)
	return rule, nil
}

func (m *Manager) DeleteRule(id, actor string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return false, errors.New("risk manager closed")
	}
	for i, v := range m.rules {
		if v.ID == id {
			previous := append([]Rule(nil), m.rules...)
			m.rules = append(m.rules[:i], m.rules[i+1:]...)
			m.generation++
			if err := m.saveLocked(); err != nil {
				m.rules = previous
				return false, err
			}
			m.recordEventLocked(v.CIDR, "", "rule_deleted", actor+": "+v.Reason)
			return true, nil
		}
	}
	return false, nil
}

func (m *Manager) History(limit int) ([]Event, []Rule) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.trimEventsLocked()
	if limit < 1 || limit > 200 {
		limit = 100
	}
	from := len(m.events) - limit
	if from < 0 {
		from = 0
	}
	events := make([]Event, len(m.events)-from)
	for i := range events {
		events[i] = m.events[len(m.events)-i-1]
	}
	rules := append([]Rule(nil), m.rules...)
	return events, rules
}
