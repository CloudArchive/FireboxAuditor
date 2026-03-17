package main

import (
	"fmt"
	"strings"
)

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
	DeviceInfo DeviceInfo    `json:"device_info"`
	Score      int           `json:"score"`
	Results    []AuditResult `json:"results"`
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
	return AuditReport{
		DeviceInfo: ExtractDeviceInfo(cfg),
		Score:      score,
		Results:    results,
	}
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

	// Map proxy actions for quick lookup
	proxyMap := make(map[string]ProxyAction)
	for _, pa := range cfg.ProxyActionList.ProxyActions {
		proxyMap[pa.Name] = pa
	}

	for i, policy := range cfg.PolicyList.Policies {
		if policy.Enabled == "0" || policy.Enabled == "false" {
			continue
		}

		var missing []string

		// 1. Check IPS (Directly on policy)
		if policy.IPSMonitor != "1" && policy.IPSMonitor != "true" {
			missing = append(missing, "IPS")
		}

		// 2. Check Proxy Services (GAV, WebBlocker, APT Blocker)
		if policy.Proxy != "" {
			if pa, ok := proxyMap[policy.Proxy]; ok {
				var gav, wb, apt string
				if pa.HTTP != nil {
					gav, wb, apt = pa.HTTP.GatewayAV, pa.HTTP.WebBlocker, pa.HTTP.APTBlocker
				} else if pa.HTTPS != nil {
					gav, wb, apt = pa.HTTPS.GatewayAV, pa.HTTPS.WebBlocker, pa.HTTPS.APTBlocker
				} else if pa.TCP != nil {
					gav, apt = pa.TCP.GatewayAV, pa.TCP.APTBlocker
				}

				if gav != "1" && gav != "true" {
					missing = append(missing, "Gateway AntiVirus")
				}
				if wb != "1" && wb != "true" {
					if pa.TCP == nil { // TCP proxy doesn't have WebBlocker
						missing = append(missing, "WebBlocker")
					}
				}
				if apt != "1" && apt != "true" {
					missing = append(missing, "APT Blocker")
				}
			}
		}

		if len(missing) > 0 {
			msg := fmt.Sprintf("[%d] %s: Eksik servisler (%s)", i+1, policy.Name, strings.Join(missing, ", "))
			r.Details = append(r.Details, msg)
		}
	}

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
