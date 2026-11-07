package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestServerRiskOptInAndInitializationFailure(t *testing.T) {
	t.Setenv("CAREOLD_RISK_DATA_DIR", "")
	if m, err := openServerRisk(); m != nil || err != nil {
		t.Fatal("risk enabled without opt-in")
	}
	t.Setenv("CAREOLD_RISK_DATA_DIR", t.TempDir())
	t.Setenv("CAREOLD_RISK_NETWORK_FEED", "")
	t.Setenv("CAREOLD_RISK_PROXY_CIDRS_FILE", "")
	m, err := openServerRisk()
	if err != nil {
		t.Fatal(err)
	}
	m.FailedLogin("192.0.2.1", "13800000001")
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(os.Getenv("CAREOLD_RISK_DATA_DIR"), "risk-security.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatal("shutdown snapshot absent", err)
	}
	if err := os.WriteFile(path, []byte("invalid JSON"), 0600); err != nil {
		t.Fatal(err)
	}
	if m, err := openServerRisk(); err == nil || m != nil {
		t.Fatal("corrupt policy silently enabled an empty manager")
	}
}
