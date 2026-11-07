package main

import (
	"os"
	"strings"

	"suining-care/internal/risk"
)

// Opt in explicitly so a desktop upgrade does not silently block existing users.
// The directory must have exactly one server writer. SQL login_audit remains
// the authoritative authentication audit; JSON contains supplementary signals.
func openServerRisk() (*risk.Manager, error) {
	dir := strings.TrimSpace(os.Getenv("CAREOLD_RISK_DATA_DIR"))
	if dir == "" {
		return nil, nil
	}
	m := risk.New()
	if path := os.Getenv("CAREOLD_RISK_NETWORK_FEED"); path != "" {
		if err := m.LoadNetworkSignals(path); err != nil {
			return nil, err
		}
	}
	if path := os.Getenv("CAREOLD_RISK_PROXY_CIDRS_FILE"); path != "" {
		if err := m.LoadProxyCIDRs(path); err != nil {
			return nil, err
		}
	}
	if err := m.Persist(dir); err != nil {
		return nil, err
	}
	return m, nil
}
