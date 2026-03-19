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
	Policies   []Policy      `json:"policies"`
	Aliases    []Alias       `json:"aliases"`
}

func RunAudit(cfg *WatchGuardConfig) AuditReport {
	results := []AuditResult{
		checkDefaultPasswords(cfg),
		checkManagementExposure(cfg),
		checkOutgoingPolicy(cfg),
		checkSecurityServices(cfg),
		checkLogging(cfg),
	}

	// Map proxy actions for enrichment
	proxyMap := make(map[string]ProxyAction)
	for _, pa := range cfg.ProxyActionList.ProxyActions {
		proxyMap[pa.Name] = pa
	}

	// Assign order, populate log settings and proxy services
	for i := range cfg.PolicyList.Policies {
		cfg.PolicyList.Policies[i].Order = i + 1

		p := &cfg.PolicyList.Policies[i]

		// Populate LogSettings from raw sibling XML fields
		p.Logging = LogSettings{
			Enabled:   p.LogRaw,
			ForReport: p.LogForReport,
		}

		if p.Proxy != "" {
			if pa, ok := proxyMap[p.Proxy]; ok {
				p.ProxyServices = resolveProxyServices(pa, proxyMap, p.IPSMonitor)
			}
		}
	}

	// Build address-group lookup for resolving IP addresses
	addrGroups := make(map[string]AddressGroup)
	for _, ag := range cfg.AddressGroupList {
		addrGroups[ag.Name] = ag
	}

	// Merge alias sources and resolve members
	allAliases := append(cfg.PolicyObjects.Aliases, cfg.AliasList...)
	for i := range allAliases {
		resolveAliasMembers(&allAliases[i], addrGroups)
	}

	score := calculateScore(results)
	return AuditReport{
		DeviceInfo: ExtractDeviceInfo(cfg),
		Score:      score,
		Results:    results,
		Policies:   cfg.PolicyList.Policies,
		Aliases:    allAliases,
	}
}

// hasHTTPRedirect reports whether a TCP proxy action redirects any traffic to an HTTP/HTTPS sub-proxy.
func hasHTTPRedirect(tcp *TCPProxyAction) bool {
	for _, r := range tcp.Redirects {
		if r.Pattern == "http" || r.Pattern == "ssl" {
			return true
		}
	}
	return false
}

// resolveProxyServices extracts security service flags for a policy's proxy action.
// For TCP-UDP proxy actions it follows the redirect rules to HTTP/HTTPS sub-actions
// and OR-combines their security service flags (if any sub-action has it enabled → true).
func resolveProxyServices(pa ProxyAction, proxyMap map[string]ProxyAction, ipsMonitor string) *PolicyProxyServices {
	ps := &PolicyProxyServices{
		IPS: ipsMonitor == "1" || ipsMonitor == "true",
	}

	boolVal := func(s string) bool { return s == "1" || s == "true" }

	setWebBlocker := func(profile string) {
		if profile != "" {
			ps.WebBlocker = true
			if ps.WebBlockerProfile == "" {
				ps.WebBlockerProfile = profile
			}
		}
	}

	switch {
	case pa.HTTP != nil:
		ps.GatewayAV = boolVal(pa.HTTP.GatewayAV)
		setWebBlocker(pa.HTTP.WebBlocker)
		ps.APTBlocker = boolVal(pa.HTTP.APTBlocker)

	case pa.HTTPS != nil:
		// HTTPS proxy redirects content inspection to an HTTP proxy action.
		// Security services (GAV, APT, WebBlocker) live on that HTTP action.
		if pa.HTTPS.RedirectTo != "" {
			if sub, ok := proxyMap[pa.HTTPS.RedirectTo]; ok && sub.HTTP != nil {
				ps.GatewayAV = boolVal(sub.HTTP.GatewayAV)
				ps.APTBlocker = boolVal(sub.HTTP.APTBlocker)
				setWebBlocker(sub.HTTP.WebBlocker)
			}
		}
		// wb-inspect on HTTPS itself also signals WebBlocker
		if boolVal(pa.HTTPS.WebBlockerInspect) {
			ps.WebBlocker = true
		}

	case pa.TCP != nil:
		// TCP-UDP proxy redirects HTTP/HTTPS traffic to dedicated sub-proxy actions.
		// Resolve each redirect and merge security service flags.
		for _, rule := range pa.TCP.Redirects {
			sub, ok := proxyMap[rule.Name]
			if !ok {
				continue
			}
			switch {
			case sub.HTTP != nil:
				if boolVal(sub.HTTP.GatewayAV) {
					ps.GatewayAV = true
				}
				setWebBlocker(sub.HTTP.WebBlocker)
				if boolVal(sub.HTTP.APTBlocker) {
					ps.APTBlocker = true
				}
			case sub.HTTPS != nil:
				// HTTPS sub-action — follow its redirect-to as well
				if sub.HTTPS.RedirectTo != "" {
					if httpSub, ok2 := proxyMap[sub.HTTPS.RedirectTo]; ok2 && httpSub.HTTP != nil {
						if boolVal(httpSub.HTTP.GatewayAV) {
							ps.GatewayAV = true
						}
						setWebBlocker(httpSub.HTTP.WebBlocker)
						if boolVal(httpSub.HTTP.APTBlocker) {
							ps.APTBlocker = true
						}
					}
				}
				if boolVal(sub.HTTPS.WebBlockerInspect) {
					ps.WebBlocker = true
				}
			}
		}
	}

	return ps
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

		// 1. Check IPS (directly on policy)
		if policy.IPSMonitor != "1" && policy.IPSMonitor != "true" {
			missing = append(missing, "IPS")
		}

		// 2. Check proxy security services (follows TCP-UDP redirect chains)
		if policy.Proxy != "" {
			if pa, ok := proxyMap[policy.Proxy]; ok {
				ps := resolveProxyServices(pa, proxyMap, policy.IPSMonitor)
				if !ps.GatewayAV {
					missing = append(missing, "Gateway AntiVirus")
				}
				// WebBlocker only applies when there's an HTTP/HTTPS proxy action reachable
				hasHTTP := pa.HTTP != nil || pa.HTTPS != nil ||
					(pa.TCP != nil && hasHTTPRedirect(pa.TCP))
				if hasHTTP && !ps.WebBlocker {
					missing = append(missing, "WebBlocker")
				}
				if !ps.APTBlocker {
					missing = append(missing, "APT Blocker")
				}
			}
		}

		if len(missing) > 0 {
			msg := fmt.Sprintf("[%d] %s → %s", i+1, policy.Name, strings.Join(missing, ", "))
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
			policy.Logging.ForReport != "true" && policy.Logging.ForReport != "1" {
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
		findings := len(r.Details)
		if findings == 0 {
			findings = 1 // At least 1 finding if rule failed
		}
		switch r.Severity {
		case Critical:
			score -= min(findings*10, 30) // 10 per finding, max 30
		case High:
			score -= min(findings*3, 20) // 3 per finding, max 20
		case Medium:
			score -= min(findings*2, 10) // 2 per finding, max 10
		}
	}
	if score < 0 {
		score = 0
	}
	return score
}
