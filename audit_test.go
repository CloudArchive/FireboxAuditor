package main

import (
	"os"
	"testing"
)

func TestAuditSampleConfig(t *testing.T) {
	data, err := os.ReadFile("testdata/sample-config.xml")
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := ParseConfig(data)
	if err != nil {
		t.Fatal(err)
	}
	report := RunAudit(cfg)

	// Sample config has: default passwords (-20), mgmt exposure (-20), outgoing active (-10),
	// IPS+APT disabled (-10), outgoing no logging (-5) = 100-65 = 35
	if report.Score > 50 {
		t.Errorf("expected low score for insecure config, got %d", report.Score)
	}

	failCount := 0
	for _, r := range report.Results {
		if !r.Passed {
			failCount++
			t.Logf("FAIL: %s (%s) - %s", r.Name, r.Severity, r.Description)
		}
	}
	if failCount == 0 {
		t.Error("expected at least some failures")
	}
}
