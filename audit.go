package main

import "strings"

type Severity string

const (
	Critical Severity = "critical"
	High     Severity = "high"
	Medium   Severity = "medium"
)

type AuditResult struct {
	RuleID      string   `json:"rule_id"`
	Name        string   `json:"name"`
	Severity    Severity `json:"severity"`
	Passed      bool     `json:"passed"`
	Description string   `json:"description"`
	Remediation string   `json:"remediation"`
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
		RuleID:      "RULE-001",
		Name:        "Varsayılan Şifre Kontrolü",
		Severity:    Critical,
		Passed:      true,
		Description: "Tüm yönetici hesaplarının şifreleri fabrika ayarından değiştirilmiş.",
		Remediation: "Firebox System Manager veya Web UI üzerinden admin ve status hesaplarının şifrelerini güçlü, benzersiz parolalarla değiştirin.",
	}

	defaults := map[string]string{
		"admin":  "readwrite",
		"status": "readonly",
	}

	var found []string
	for _, user := range cfg.SystemParameters.AdminUsers {
		if expected, ok := defaults[strings.ToLower(user.Name)]; ok {
			if strings.EqualFold(user.Password, expected) || user.Password == "" {
				found = append(found, user.Name)
			}
		}
	}

	if len(found) > 0 {
		r.Passed = false
		r.Description = "Varsayılan şifre kullanan hesaplar tespit edildi: " + strings.Join(found, ", ") + ". Bu, cihazınıza yetkisiz erişim riski oluşturur."
	}
	return r
}

// Rule 2 (Critical): Management interface exposure
func checkManagementExposure(cfg *WatchGuardConfig) AuditResult {
	r := AuditResult{
		RuleID:      "RULE-002",
		Name:        "Yönetim Arayüzü İfşası",
		Severity:    Critical,
		Passed:      true,
		Description: "Yönetim arayüzleri dış ağdan erişime kapalı.",
		Remediation: "WatchGuard ve WatchGuard Web UI politikalarının 'From' alanından 'Any-External' objesini kaldırın. Yönetim erişimini yalnızca güvenilir iç ağ IP'leriyle sınırlandırın.",
	}

	mgmtPolicies := []string{"watchguard", "watchguard web ui"}
	var exposed []string

	for _, policy := range cfg.PolicyList.Policies {
		nameLower := strings.ToLower(policy.Name)
		for _, mgmt := range mgmtPolicies {
			if nameLower == mgmt {
				for _, alias := range policy.From.Aliases {
					if strings.EqualFold(alias, "Any-External") {
						exposed = append(exposed, policy.Name)
					}
				}
			}
		}
	}

	if len(exposed) > 0 {
		r.Passed = false
		r.Description = "Dış ağdan erişime açık yönetim politikaları: " + strings.Join(exposed, ", ") + ". Saldırganlar cihazınızın yönetim paneline internet üzerinden erişebilir."
	}
	return r
}

// Rule 3 (High): Outgoing policy
func checkOutgoingPolicy(cfg *WatchGuardConfig) AuditResult {
	r := AuditResult{
		RuleID:      "RULE-003",
		Name:        "Outgoing (Giden Trafik) Politikası",
		Severity:    High,
		Passed:      true,
		Description: "Varsayılan 'Outgoing' politikası devre dışı veya kaldırılmış.",
		Remediation: "Varsayılan 'Outgoing' politikasını devre dışı bırakın ve iç ağdan dışarıya yalnızca gerekli port/protokollere izin veren özel kurallar oluşturun (HTTP, HTTPS, DNS vb.).",
	}

	for _, policy := range cfg.PolicyList.Policies {
		if strings.EqualFold(policy.Name, "Outgoing") {
			if policy.Enabled != "false" && policy.Enabled != "0" {
				r.Passed = false
				r.Description = "'Outgoing' politikası aktif durumda. Bu politika, iç ağdan dışarıya tüm TCP/UDP trafiğine izin verir ve veri sızıntısı, kötü amaçlı yazılım iletişimi gibi risklere kapı açar."
			}
			break
		}
	}
	return r
}

// Rule 4 (High): Security services
func checkSecurityServices(cfg *WatchGuardConfig) AuditResult {
	r := AuditResult{
		RuleID:      "RULE-004",
		Name:        "Güvenlik Servisleri Durumu",
		Severity:    High,
		Passed:      true,
		Description: "Gateway AV, IPS, WebBlocker ve APT Blocker servisleri aktif.",
		Remediation: "Firebox System Manager > Subscription Services bölümünden Gateway AntiVirus, IPS, WebBlocker ve APT Blocker'ı etkinleştirin. Lisans durumunuzu kontrol edin.",
	}

	var disabled []string
	ss := cfg.SecurityServices

	check := func(name string, svc *ServiceGlobal) {
		if svc == nil || (svc.Enabled != "true" && svc.Enabled != "1") {
			disabled = append(disabled, name)
		}
	}

	check("Gateway AntiVirus", ss.GatewayAV)
	check("IPS", ss.IPS)
	check("WebBlocker", ss.WebBlocker)
	check("APT Blocker", ss.APTBlocker)

	if len(disabled) > 0 {
		r.Passed = false
		r.Description = "Devre dışı güvenlik servisleri: " + strings.Join(disabled, ", ") + ". Bu servisler olmadan cihazınız zararlı yazılım, saldırı ve tehlikeli web sitelerine karşı korumasız kalır."
	}
	return r
}

// Rule 5 (Medium): Logging
func checkLogging(cfg *WatchGuardConfig) AuditResult {
	r := AuditResult{
		RuleID:      "RULE-005",
		Name:        "Politika Loglama Kontrolü",
		Severity:    Medium,
		Passed:      true,
		Description: "Tüm politikalarda loglama aktif.",
		Remediation: "Policy Manager'da her bir politikayı düzenleyip 'Properties' > 'Logging' sekmesinden 'Send a log message' ve 'Send a log message for reports' seçeneklerini işaretleyin.",
	}

	var unlogged []string
	for _, policy := range cfg.PolicyList.Policies {
		if policy.Enabled == "false" || policy.Enabled == "0" {
			continue
		}
		if policy.Logging.Enabled != "true" && policy.Logging.Enabled != "1" &&
			policy.Logging.ForReport != "true" && policy.Logging.ForReport != "1" &&
			policy.Logging.LogMessage != "true" && policy.Logging.LogMessage != "1" {
			unlogged = append(unlogged, policy.Name)
		}
	}

	if len(unlogged) > 0 {
		r.Passed = false
		r.Description = "Loglama kapalı olan politikalar: " + strings.Join(unlogged, ", ") + ". Loglar olmadan güvenlik olaylarını tespit etmek ve adli analiz yapmak imkansız hale gelir."
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
