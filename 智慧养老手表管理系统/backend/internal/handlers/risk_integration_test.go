package handlers

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"suining-care/internal/auth"
	"suining-care/internal/config"
	"suining-care/internal/csrf"
	"suining-care/internal/risk"
	"suining-care/internal/securestore"
	"suining-care/internal/session"
)

func TestRiskGatesBeforeDatabaseAccess(t *testing.T) {
	store := session.NewStore(time.Hour)
	t.Cleanup(store.Close)
	h := New(nil, store, config.Config{})
	h.Risk = risk.New()
	if _, err := h.Risk.AddRule("192.0.2.1", "block", "confirmed abuse", "admin", time.Hour); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		path   string
		action http.HandlerFunc
		status int
	}{
		{"/api/v1/auth/login", h.login, 401},
		{"/api/v1/inviteCode/checkCode", h.inviteCheckCode, 404},
		{"/api/v1/inviteCode/join", h.inviteJoin, 404},
	} {
		r := httptest.NewRequest("POST", tc.path, strings.NewReader(`{"phone":"13800000001","password":"Password123!"}`))
		r.RemoteAddr = "192.0.2.1:1234"
		w := httptest.NewRecorder()
		tc.action(w, r)
		if w.Code != tc.status {
			t.Fatalf("%s: HTTP %d", tc.path, w.Code)
		}
		if strings.Contains(w.Body.String(), "confirmed abuse") || strings.Contains(w.Body.String(), "retryAfter") {
			t.Fatal("risk policy leaked")
		}
	}
}

func TestRiskManagementRequiresAdminAndCSRF(t *testing.T) {
	store := session.NewStore(time.Hour)
	t.Cleanup(store.Close)
	h := New(nil, store, config.Config{})
	h.Risk = risk.New()
	mux := http.NewServeMux()
	h.routeRisk(mux)
	secured := csrf.Middleware(store, auth.Middleware(store, mux))
	for _, tc := range []struct {
		role   string
		csrf   bool
		status int
	}{
		{"", true, 403}, {auth.RoleFamily, true, 403}, {auth.RoleAdmin, false, 403}, {auth.RoleAdmin, true, 200},
	} {
		s := store.Create(session.Principal{Role: tc.role, ID: 1})
		r := httptest.NewRequest("POST", "/api/v1/risk/rules", strings.NewReader(`{"ip":"192.0.2.3","action":"block","reason":"confirmed abuse","durationHours":1}`))
		r.AddCookie(&http.Cookie{Name: session.CookieName, Value: s.ID})
		if tc.csrf {
			r.Header.Set("X-CSRF-TOKEN", s.CSRFToken)
		}
		w := httptest.NewRecorder()
		secured.ServeHTTP(w, r)
		if w.Code != tc.status {
			t.Fatalf("role=%q csrf=%v: HTTP %d %s", tc.role, tc.csrf, w.Code, w.Body.String())
		}
	}
	_, rules := h.Risk.History(100)
	if len(rules) != 1 || rules[0].Actor != "admin:1" {
		t.Fatalf("actor attribution: %#v", rules)
	}
}

func TestRiskAuthenticationAuditPersistence(t *testing.T) {
	pool := testDatabase(t)
	h := testHandler(t, pool)
	box, err := securestore.New(base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x42}, 32)))
	if err != nil {
		t.Fatal(err)
	}
	h.Box = box
	h.Risk = risk.New()
	dir := t.TempDir()
	if err := h.Risk.Persist(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := h.Risk.Close(); err != nil {
			t.Error(err)
		}
	})
	hash, err := auth.Hash("TestPassword123!")
	if err != nil {
		t.Fatal(err)
	}
	mustExec(t, pool, "INSERT INTO users(user_id,user_name,user_phone,password,role) VALUES(1,'test','13800000001',$1,'1')", hash)
	for i := 0; i < 5; i++ {
		w := callAPI(h.login, map[string]any{"phone": "13800000001", "password": "WrongPassword123"})
		if w.Code != 401 {
			t.Fatalf("failure %d: HTTP %d", i, w.Code)
		}
	}
	blocked := callAPI(h.login, map[string]any{"phone": "13800000001", "password": "TestPassword123!"})
	if blocked.Code != 401 || len(blocked.Result().Cookies()) != 0 {
		t.Fatal("risk block did not prevent session creation")
	}
	// A manual exception bypasses only the automatic risk penalty, never SQL audit or password checks.
	if _, err := h.Risk.AddRule("192.0.2.1", "skip-auto", "trusted test peer", "admin", time.Hour); err != nil {
		t.Fatal(err)
	}
	w := callAPI(h.login, map[string]any{"phone": "13800000001", "password": "TestPassword123!"})
	if w.Code != 200 || len(w.Result().Cookies()) != 1 {
		t.Fatalf("login after exception: HTTP %d %s", w.Code, w.Body.String())
	}
	current := h.Sessions.Get(w.Result().Cookies()[0].Value)
	if current == nil || current.Principal.Role != auth.RoleAdmin || current.Principal.ID != 1 {
		t.Fatal("risk exception did not establish the existing authenticated principal")
	}
	if err := h.Risk.Close(); err != nil {
		t.Fatal(err)
	}
	fresh := risk.New()
	if err := fresh.Persist(dir); err != nil {
		t.Fatal(err)
	}
	defer fresh.Close()
	events, _ := fresh.History(200)
	if len(events) != 8 || events[0].Type != "login_success" {
		t.Fatalf("unexpected persisted audit: %#v", events)
	}
	// The legacy PG compatibility path intentionally has no login_audit table;
	// loginAuditTx is MySQL-only. Keep that backend contract unchanged.
	if pool.IsMySQL() {
		var count int
		if err := pool.QueryRow(t.Context(), "SELECT COUNT(*) FROM login_audit WHERE outcome='success'").Scan(&count); err != nil || count != 1 {
			t.Fatalf("authoritative MySQL audit lost: %d %v", count, err)
		}
	}
}
