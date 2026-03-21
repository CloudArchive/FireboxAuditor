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
	RuleID         string   `json:"rule_id"`
	Severity       Severity `json:"severity"`
	Passed         bool     `json:"passed"`
	Details        []string `json:"details,omitempty"`
	PointsDeducted int      `json:"points_deducted"`
}

type AuditReport struct {
	DeviceInfo       DeviceInfo            `json:"device_info"`
	Score            int                   `json:"score"`
	Results          []AuditResult         `json:"results"`
	Policies         []Policy              `json:"policies"`
	Aliases          []Alias               `json:"aliases"`
	NATRules         []NATRule             `json:"nat_rules,omitempty"`
	GlobalSecurity   *GlobalSecurityStatus `json:"global_security,omitempty"`
}

// GlobalSecurityStatus summarizes global security feature states.
type GlobalSecurityStatus struct {
	IPSEnabled             bool   `json:"ips_enabled"`
	IPSFullScan            bool   `json:"ips_full_scan"`
	APTEnabled             bool   `json:"apt_enabled"`
	BotnetDetection        bool   `json:"botnet_detection"`
	DoSPreventionEnabled   bool   `json:"dos_prevention_enabled"`
	DoSEnabledRuleCount    int    `json:"dos_enabled_rule_count"`
	DoSTotalRuleCount      int    `json:"dos_total_rule_count"`
}

func RunAudit(cfg *WatchGuardConfig) AuditReport {
	results := []AuditResult{
		checkDefaultPasswords(cfg),
		checkManagementExposure(cfg),
		checkOutgoingPolicy(cfg),
		checkSecurityServices(cfg),
		checkLogging(cfg),
		checkOrphanObjects(cfg),
		checkAccountLockout(cfg),
		checkVPNWeakCrypto(cfg),
		checkDefaultCertificates(cfg),
	}

	// Map proxy actions for enrichment
	proxyMap := make(map[string]ProxyAction)
	for _, pa := range cfg.ProxyActionList.ProxyActions {
		proxyMap[pa.Name] = pa
	}

	// Build service lookup map
	svcMap := make(map[string]ServiceDef)
	for _, svc := range cfg.ServiceList {
		svcMap[svc.Name] = svc
	}

	// Build NAT lookup map
	natMap := make(map[string]NATRule)
	for _, nr := range cfg.NATList {
		natMap[nr.Name] = nr
	}

	// Build address-group lookup for resolving IP addresses
	addrGroups := make(map[string]AddressGroup)
	for _, ag := range cfg.AddressGroupList {
		addrGroups[ag.Name] = ag
	}

	// Assign order, populate log settings, service ports, NAT details, and proxy services
	for i := range cfg.PolicyList.Policies {
		cfg.PolicyList.Policies[i].Order = i + 1

		p := &cfg.PolicyList.Policies[i]
		p.IsSystem = isSystemPolicy(p.Name)

		// Populate LogSettings from raw sibling XML fields
		p.Logging = LogSettings{
			Enabled:   p.LogRaw,
			ForReport: p.LogForReport,
		}

		// Resolve service name to actual ports/protocols
		if svc, ok := svcMap[p.Service]; ok {
			p.ServicePorts = ServicePortSummary(svc)
		}

		// Resolve NAT reference to actual NAT rule details
		if p.NATRef != "" {
			if nr, ok := natMap[p.NATRef]; ok {
				p.ResolvedNAT = resolveNATRule(nr, addrGroups)
			}
		}

		if p.Proxy != "" {
			if pa, ok := proxyMap[p.Proxy]; ok {
				p.ProxyServices = resolveProxyServices(pa, proxyMap, p.IPSMonitor)
			}
		}
	}

	// Merge alias sources and resolve members
	allAliases := append(cfg.PolicyObjects.Aliases, cfg.AliasList...)
	for i := range allAliases {
		resolveAliasMembers(&allAliases[i], addrGroups)
	}

	// Extract global security status
	globalSec := extractGlobalSecurity(cfg)

	score := calculateScore(results)
	return AuditReport{
		DeviceInfo:     ExtractDeviceInfo(cfg),
		Score:          score,
		Results:        results,
		Policies:       cfg.PolicyList.Policies,
		Aliases:        allAliases,
		NATRules:       cfg.NATList,
		GlobalSecurity: globalSec,
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

// resolveNATRule creates a resolved NAT summary, following addr-name references
// to address-groups to extract actual IPs. Produces members like "Any-External --> 10.168.1.250".
func resolveNATRule(nr NATRule, addrGroups map[string]AddressGroup) *ResolvedNAT {
	rn := &ResolvedNAT{
		Name:     nr.Name,
		TypeName: NATTypeName(nr.Type),
	}
	for _, m := range nr.Items {
		iface := m.Interface
		if iface == "" {
			iface = m.ExtAddrName
		}

		// Resolve addr-name to actual IP via address-group lookup
		targetIP := m.IP
		if targetIP == "" && m.AddrName != "" {
			if ag, ok := addrGroups[m.AddrName]; ok {
				for _, gm := range ag.Members {
					if s := resolveAddressGroupMember(gm); s != "" {
						targetIP = s
						break
					}
				}
			}
		}

		if iface != "" && targetIP != "" {
			rn.Members = append(rn.Members, fmt.Sprintf("%s --> %s", iface, targetIP))
		} else if targetIP != "" {
			rn.Members = append(rn.Members, targetIP)
		} else if iface != "" {
			rn.Members = append(rn.Members, iface)
		}
	}
	return rn
}

// extractGlobalSecurity reads global security toggles from system-parameters.
func extractGlobalSecurity(cfg *WatchGuardConfig) *GlobalSecurityStatus {
	boolVal := func(s string) bool { return s == "1" || s == "true" }

	gs := &GlobalSecurityStatus{}

	if cfg.SystemParameters.IPS != nil {
		gs.IPSEnabled = boolVal(cfg.SystemParameters.IPS.Enabled)
		gs.IPSFullScan = boolVal(cfg.SystemParameters.IPS.FullScanMode)
	}
	if cfg.SystemParameters.APT != nil {
		gs.APTEnabled = boolVal(cfg.SystemParameters.APT.Enabled)
	}
	if cfg.SystemParameters.BotnetDetection != nil {
		gs.BotnetDetection = boolVal(cfg.SystemParameters.BotnetDetection.Enabled)
	}

	gs.DoSTotalRuleCount = len(cfg.SystemParameters.DoSPrevention)
	for _, d := range cfg.SystemParameters.DoSPrevention {
		if d.Enabled == 1 {
			gs.DoSEnabledRuleCount++
		}
	}
	gs.DoSPreventionEnabled = gs.DoSEnabledRuleCount > 0

	return gs
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

	// Map proxy actions for lookup
	proxyMap := make(map[string]ProxyAction)
	for _, pa := range cfg.ProxyActionList.ProxyActions {
		proxyMap[pa.Name] = pa
	}

	for _, policy := range cfg.PolicyList.Policies {
		if policy.Enabled == "false" || policy.Enabled == "0" {
			continue
		}
		if strings.TrimSpace(policy.Name) == "" {
			continue
		}

		logRawEnabled := policy.Logging.Enabled == "true" || policy.Logging.Enabled == "1"
		logReportEnabled := policy.Logging.ForReport == "true" || policy.Logging.ForReport == "1"

		if policy.Proxy != "" {
			// Scenario C - Proxy Policies
			if pa, ok := proxyMap[policy.Proxy]; ok {
				proxyReportEnabled := pa.LogForReport == "true" || pa.LogForReport == "1"
				if !proxyReportEnabled {
					r.Details = append(r.Details, fmt.Sprintf("%s (Missing Report Logging in ProxyAction)", policy.Name))
				}
			}
		} else {
			// Scenario A and B - Packet Filter
			actionLower := strings.ToLower(policy.Action)
			isDeny := actionLower == "deny" || actionLower == "drop" || actionLower == "reject" || actionLower == "0"

			if isDeny {
				// Scenario B - Packet Filter (Action = Deny)
				if !logRawEnabled {
					r.Details = append(r.Details, fmt.Sprintf("%s (Deny policy is not logging blocked traffic)", policy.Name))
				}
			} else {
				// Scenario A - Packet Filter (Action = Allow)
				if logRawEnabled {
					r.Details = append(r.Details, fmt.Sprintf("%s (Performance Risk: Send Log Message enabled on an Allow policy)", policy.Name))
				}
				if !logReportEnabled {
					r.Details = append(r.Details, fmt.Sprintf("%s (Missing Report Logging)", policy.Name))
				}
			}
		}
	}

	if len(r.Details) > 0 {
		r.Passed = false
	}
	return r
}

func calculateScore(results []AuditResult) int {
	score := 100
	for i := range results {
		r := &results[i]
		if r.Passed {
			continue
		}
		findings := len(r.Details)
		if findings == 0 {
			findings = 1 // At least 1 finding if rule failed
		}

		penalty := 0
		switch r.Severity {
		case Critical:
			penalty = min(20+(5*findings), 30) // Base penalty 20 + (5 * findings). Max penalty: 30
		case High:
			penalty = min(10+(2*findings), 20) // Base penalty 10 + (2 * findings). Max penalty: 20
		case Medium:
			penalty = min(5+(1*findings), 10) // Base penalty 5 + (1 * findings). Max penalty: 10
		}

		r.PointsDeducted = penalty
		score -= penalty
	}
	if score < 0 {
		score = 0
	}
	return score
}

// ── Rule 6 (Medium): Orphan Objects ─────────────────────────────────────────

// isSystemDefaultAlias returns true for WatchGuard built-in aliases that cannot
// be deleted and should not be flagged as orphans.
func isSystemDefaultAlias(name string) bool {
	defaults := map[string]bool{
		"Any":            true,
		"Any-Trusted":    true,
		"Any-External":   true,
		"Any-Optional":   true,
		"Any-BOVPN":      true,
		"Any-MUVPN":      true,
		"Firebox":        true,
	}
	return defaults[name]
}

func checkOrphanObjects(cfg *WatchGuardConfig) AuditResult {
	r := AuditResult{
		RuleID:   "RULE-006",
		Severity: Medium,
		Passed:   true,
	}

	// Collect all known object names
	allAliases := append(cfg.PolicyObjects.Aliases, cfg.AliasList...)
	objectNames := make(map[string]bool)
	for _, a := range allAliases {
		objectNames[a.Name()] = true
	}
	for _, ag := range cfg.AddressGroupList {
		objectNames[ag.Name] = true
	}

	// Build "used" set from policy From/To aliases
	usedSet := make(map[string]bool)
	for _, policy := range cfg.PolicyList.Policies {
		for _, alias := range policy.From.Aliases {
			usedSet[alias] = true
		}
		for _, alias := range policy.To.Aliases {
			usedSet[alias] = true
		}
	}

	// Also mark objects referenced inside other aliases (inter-alias references)
	for _, a := range allAliases {
		for _, m := range a.RawMembers {
			if m.AliasName != "" {
				usedSet[m.AliasName] = true
			}
			if m.Address != "" {
				usedSet[m.Address] = true
			}
		}
	}

	// Check which objects are never referenced
	for name := range objectNames {
		if isSystemDefaultAlias(name) {
			continue
		}
		if !usedSet[name] {
			r.Details = append(r.Details, name)
		}
	}

	if len(r.Details) > 0 {
		r.Passed = false
	}
	return r
}

// ── Rule 7 (High): Account Lockout Settings ─────────────────────────────────

func checkAccountLockout(cfg *WatchGuardConfig) AuditResult {
	r := AuditResult{
		RuleID:   "RULE-007",
		Severity: High,
		Passed:   true,
	}

	if cfg.SystemParameters.AuthGlobal == nil ||
		cfg.SystemParameters.AuthGlobal.MgmtAcctLockout == nil {
		r.Passed = false
		r.Details = append(r.Details, "Account lockout is not configured")
		return r
	}

	lock := cfg.SystemParameters.AuthGlobal.MgmtAcctLockout
	if lock.Enabled != "1" && lock.Enabled != "true" {
		r.Passed = false
		r.Details = append(r.Details, "Account lockout is disabled")
	}
	return r
}

// ── Rule 8 (Critical): VPN Weak Cryptography ────────────────────────────────

// ikeEncryptionName maps WatchGuard encryp-algm IDs to human-readable names.
func ikeEncryptionName(algm int) string {
	switch algm {
	case 1:
		return "DES-CBC"
	case 2:
		return "DES"
	case 3:
		return "3DES"
	case 7:
		return "AES-CBC"
	case 12:
		return "AES"
	case 20:
		return "AES-GCM"
	default:
		return fmt.Sprintf("Alg-%d", algm)
	}
}

// ikeAuthName maps WatchGuard auth-algm IDs to human-readable names.
func ikeAuthName(algm int) string {
	switch algm {
	case 0:
		return "None"
	case 1:
		return "MD5"
	case 2:
		return "SHA1"
	case 4:
		return "SHA256"
	case 5:
		return "SHA256"
	case 6:
		return "SHA384"
	case 7:
		return "SHA512"
	default:
		return fmt.Sprintf("Auth-%d", algm)
	}
}

// dhGroupName maps DH group numbers to human-readable names.
func dhGroupName(group int) string {
	switch group {
	case 1:
		return "DH Group 1 (768-bit)"
	case 2:
		return "DH Group 2 (1024-bit)"
	case 5:
		return "DH Group 5 (1536-bit)"
	case 14:
		return "DH Group 14"
	case 19:
		return "DH Group 19 (ECP-256)"
	case 20:
		return "DH Group 20 (ECP-384)"
	default:
		return fmt.Sprintf("DH Group %d", group)
	}
}

// isWeakEncryption returns true for legacy encryption algorithms.
func isWeakEncryption(algm int) bool {
	return algm == 1 || algm == 2 || algm == 3 // DES-CBC, DES, 3DES
}

// isWeakAuth returns true for legacy authentication algorithms.
func isWeakAuth(algm int) bool {
	return algm == 1 // MD5
}

// isWeakDHGroup returns true for deprecated DH groups.
func isWeakDHGroup(group int) bool {
	return group == 1 || group == 2
}

func checkVPNWeakCrypto(cfg *WatchGuardConfig) AuditResult {
	r := AuditResult{
		RuleID:   "RULE-008",
		Severity: Critical,
		Passed:   true,
	}

	// Check Phase 1 (IKE Action) transforms
	for _, action := range cfg.IKEActionList {
		for _, m := range action.Transform.Members {
			var weakItems []string
			if isWeakEncryption(m.EncrypAlgm) {
				weakItems = append(weakItems, ikeEncryptionName(m.EncrypAlgm))
			}
			if isWeakAuth(m.AuthAlgm) {
				weakItems = append(weakItems, ikeAuthName(m.AuthAlgm))
			}
			if isWeakDHGroup(m.DHGroup) {
				weakItems = append(weakItems, dhGroupName(m.DHGroup))
			}
			if len(weakItems) > 0 {
				msg := fmt.Sprintf("Phase1 '%s' → %s", action.Name, strings.Join(weakItems, ", "))
				r.Details = append(r.Details, msg)
			}
		}
	}

	// Check Phase 2 (IPsec Proposal) transforms
	for _, prop := range cfg.IPsecProposalList {
		for _, m := range prop.ESPTransform.Members {
			var weakItems []string
			if isWeakEncryption(m.EncrypAlgm) {
				weakItems = append(weakItems, ikeEncryptionName(m.EncrypAlgm))
			}
			if isWeakAuth(m.AuthAlgm) {
				weakItems = append(weakItems, ikeAuthName(m.AuthAlgm))
			}
			if len(weakItems) > 0 {
				msg := fmt.Sprintf("Phase2 '%s' → %s", prop.Name, strings.Join(weakItems, ", "))
				r.Details = append(r.Details, msg)
			}
		}
	}

	if len(r.Details) > 0 {
		r.Passed = false
	}
	return r
}

// ── Rule 9 (Medium): Default Certificates in Use ────────────────────────────

func checkDefaultCertificates(cfg *WatchGuardConfig) AuditResult {
	r := AuditResult{
		RuleID:   "RULE-009",
		Severity: Medium,
		Passed:   true,
	}

	// Extract serial number for comparison
	serialNumber := ""
	for _, cert := range cfg.SystemParameters.IKECerts {
		if m := snRe.FindStringSubmatch(cert.Issuer); len(m) > 1 {
			serialNumber = m[1]
			break
		}
	}

	// Merge all certificate sources
	allCerts := append(cfg.SystemParameters.IKECerts, cfg.CertList...)

	for _, cert := range allCerts {
		issuerLower := strings.ToLower(cert.Issuer)
		subjectLower := strings.ToLower(cert.Subject)

		isDefault := false
		var reason string

		// Check for WatchGuard default certificate patterns
		if strings.Contains(issuerLower, "o=watchguard") ||
			strings.Contains(subjectLower, "o=watchguard") {
			isDefault = true
			reason = "WatchGuard default issuer"
		}

		if strings.Contains(issuerLower, "watchguard default") ||
			strings.Contains(subjectLower, "watchguard default") {
			isDefault = true
			reason = "WatchGuard Default certificate"
		}

		// Check if certificate contains the device serial number
		if serialNumber != "" {
			if strings.Contains(cert.Issuer, serialNumber) ||
				strings.Contains(cert.Subject, serialNumber) {
				isDefault = true
				reason = "Contains device serial number"
			}
		}

		if isDefault {
			// Truncate long subject/issuer for readability
			label := cert.Subject
			if label == "" {
				label = cert.Issuer
			}
			if len(label) > 80 {
				label = label[:80] + "..."
			}
			r.Details = append(r.Details, fmt.Sprintf("%s (%s)", label, reason))
		}
	}

	if len(r.Details) > 0 {
		r.Passed = false
	}
	return r
}

// ── Helpers ─────────────────────────────────────────────────────────────────

// isSystemPolicy helps determine if a given policy is a WatchGuard built-in system policy
func isSystemPolicy(name string) bool {
	lower := strings.ToLower(strings.TrimSpace(name))

	// Exact matches
	exactMatches := []string{
		"watchguard", "watchguard web ui", "pingtofirebox", "ping", "ntp", "dns",
		"wg-tdr-host-sensor", "watchguard authentication", "watchguard certificate portal",
		"watchguard sslvpn", "wg-auth", "wg-cert-portal", "wg-firebox-mgmt",
		"allow sslvpn-users", "allow ikev2-users", "watchguard ipsec", "dhcp-client",
		"watchguard cloud", "watchguard report server", "watchguard log server",
		"allow-ike-to-firebox",
	}
	for _, sn := range exactMatches {
		if lower == sn || lower == sn+"-00" {
			return true
		}
	}

	// Dynamic/prefix matches for auto-generated VPN and system rules
	prefixes := []string{
		"unhandled muvpn packet",
		"ipsec-vpn",
		"bovpn-allow",
		"wg-",
		"watchguard ",
	}
	for _, pfx := range prefixes {
		if strings.HasPrefix(lower, pfx) {
			return true
		}
	}

	// Auto-generated IPSec/BOVPN interface policies often end with .in, .out, .in-00, .out-00
	if strings.HasSuffix(lower, ".in") || strings.HasSuffix(lower, ".out") ||
		strings.HasSuffix(lower, ".in-00") || strings.HasSuffix(lower, ".out-00") {
		// Just to be safe, only hide them if they look like VPN policies
		if strings.Contains(lower, "ipsec") || strings.Contains(lower, "bovpn") || strings.Contains(lower, "vpn") {
			return true
		}
	}

	return false
}
