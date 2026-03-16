package main

import "strings"

type Severity string

const (
	Critical Severity = "critical"
	High     Severity = "high"
	Medium   Severity = "medium"
)

type AuditResult struct {
	RuleID   string   `json:"rule_id"`
	Severity Severity `json:"severity"`
	Passed   bool     `json:"passed"`
	Details  []string `json:"details,omitempty"`
}

type AuditReport struct {
	Score   int           `json:"score"`
	Results []AuditResult `json:"results"`
}

func RunAudit(cfg *WatchGuardConfig) AuditReport {
	results := []AuditResult{
		checkDefaultPasswords(cfg),
		checkManagementExposure(cfg),
		checkOutgoingPolicy(cfg),
		checkSecurityServices(cfg),
		checkLogging(cfg),
	}

	score := calculateScore(results)
	return AuditReport{Score: score, Results: results}
}

// Rule 1 (Critical): Default passwords
func checkDefaultPasswords(cfg *WatchGuardConfig) AuditResult {
	r := AuditResult{
		RuleID:   "RULE-001",
		Severity: Critical,
		Passed:   true,
	}

	defaults := map[string]string{
		"admin":  "readwrite",
		"status": "readonly",
	}

	for _, user := range cfg.SystemParameters.AdminUsers {
		if expected, ok := defaults[strings.ToLower(user.Name)]; ok {
			if strings.EqualFold(user.Password, expected) || user.Password == "" {
				r.Details = append(r.Details, user.Name)
			}
		}
	}

	if len(r.Details) > 0 {
		r.Passed = false
	}
	return r
}

// Rule 2 (Critical): Management interface exposure
func checkManagementExposure(cfg *WatchGuardConfig) AuditResult {
	r := AuditResult{
		RuleID:   "RULE-002",
		Severity: Critical,
		Passed:   true,
	}

	mgmtPolicies := []string{"watchguard", "watchguard web ui"}

	for _, policy := range cfg.PolicyList.Policies {
		nameLower := strings.ToLower(policy.Name)
		for _, mgmt := range mgmtPolicies {
			if nameLower == mgmt {
				for _, alias := range policy.From.Aliases {
					if strings.EqualFold(alias, "Any-External") {
						r.Details = append(r.Details, policy.Name)
					}
				}
			}
		}
	}

	if len(r.Details) > 0 {
		r.Passed = false
	}
	return r
}

// Rule 3 (High): Outgoing policy
func checkOutgoingPolicy(cfg *WatchGuardConfig) AuditResult {
	r := AuditResult{
		RuleID:   "RULE-003",
		Severity: High,
		Passed:   true,
	}

	for _, policy := range cfg.PolicyList.Policies {
		if strings.EqualFold(policy.Name, "Outgoing") {
			if policy.Enabled != "false" && policy.Enabled != "0" {
				r.Passed = false
			}
			break
		}
	}
	return r
}

// Rule 4 (High): Security services
func checkSecurityServices(cfg *WatchGuardConfig) AuditResult {
	r := AuditResult{
		RuleID:   "RULE-004",
		Severity: High,
		Passed:   true,
	}

	ss := cfg.SecurityServices

	check := func(name string, svc *ServiceGlobal) {
		if svc == nil || (svc.Enabled != "true" && svc.Enabled != "1") {
			r.Details = append(r.Details, name)
		}
	}

	check("Gateway AntiVirus", ss.GatewayAV)
	check("IPS", ss.IPS)
	check("WebBlocker", ss.WebBlocker)
	check("APT Blocker", ss.APTBlocker)

	if len(r.Details) > 0 {
		r.Passed = false
	}
	return r
}

// Rule 5 (Medium): Logging
func checkLogging(cfg *WatchGuardConfig) AuditResult {
	r := AuditResult{
		RuleID:   "RULE-005",
		Severity: Medium,
		Passed:   true,
	}

	for _, policy := range cfg.PolicyList.Policies {
		if policy.Enabled == "false" || policy.Enabled == "0" {
			continue
		}
		if strings.TrimSpace(policy.Name) == "" {
			continue
		}
		if policy.Logging.Enabled != "true" && policy.Logging.Enabled != "1" &&
			policy.Logging.ForReport != "true" && policy.Logging.ForReport != "1" &&
			policy.Logging.LogMessage != "true" && policy.Logging.LogMessage != "1" {
			r.Details = append(r.Details, policy.Name)
		}
	}

	if len(r.Details) > 0 {
		r.Passed = false
	}
	return r
}

func calculateScore(results []AuditResult) int {
	score := 100
	for _, r := range results {
		if r.Passed {
			continue
		}
		switch r.Severity {
		case Critical:
			score -= 20
		case High:
			score -= 10
		case Medium:
			score -= 5
		}
	}
	if score < 0 {
		score = 0
	}
	return score
}
