package main

import (
	"os"
	"strings"
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

	if report.Score > 50 {
		t.Errorf("expected low score for insecure config, got %d", report.Score)
	}

	failCount := 0
	for _, r := range report.Results {
		if !r.Passed {
			failCount++
			t.Logf("FAIL: %s (%s) details=%s", r.RuleID, r.Severity, strings.Join(r.Details, ", "))
		}
	}
	if failCount == 0 {
		t.Error("expected at least some failures")
	}

	// Verify no empty details entries (the comma bug)
	for _, r := range report.Results {
		for _, d := range r.Details {
			if strings.TrimSpace(d) == "" {
				t.Errorf("rule %s has empty detail entry", r.RuleID)
			}
		}
	}
}
