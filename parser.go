package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
)

// Package-level compiled regex for serial number extraction
var snRe = regexp.MustCompile(`SN\s+(\S+)`)

// wgTimeZone maps WatchGuard's internal timezone index to IANA timezone string.
// The index list is based on the WatchGuard firmware timezone enumeration.
func wgTimeZone(idx int) string {
	zones := map[int]string{
		0: "UTC-12:00 (Dateline)",
		1: "UTC-11:00 (Samoa)",
		2: "UTC-10:00 (Hawaii)",
		3: "UTC-09:00 (Alaska)",
		4: "UTC-08:00 (Pacific Time)",
		5: "UTC-07:00 (Arizona)",
		6: "UTC-07:00 (Mountain Time)",
		7: "UTC-06:00 (Central America)",
		8: "UTC-06:00 (Central Time)",
		9: "UTC-06:00 (Saskatchewan)",
		10: "UTC-06:00 (Mexico City)",
		11: "UTC-05:00 (Eastern Time)",
		12: "UTC-05:00 (Indiana)",
		13: "UTC-05:00 (Bogota/Lima)",
		14: "UTC-04:00 (Atlantic Time)",
		15: "UTC-04:00 (Caracas)",
		16: "UTC-04:00 (Santiago)",
		17: "UTC-03:30 (Newfoundland)",
		18: "UTC-03:00 (Buenos Aires)",
		19: "UTC-03:00 (Greenland)",
		20: "UTC-03:00 (Brasilia)",
		21: "UTC-02:00 (Mid-Atlantic)",
		22: "UTC-01:00 (Azores)",
		23: "UTC-01:00 (Cape Verde)",
		24: "UTC+00:00 (London/Dublin)",
		25: "UTC+00:00 (Casablanca)",
		26: "UTC+01:00 (Amsterdam/Berlin)",
		27: "UTC+01:00 (Belgrade)",
		28: "UTC+01:00 (Brussels/Paris)",
		29: "UTC+01:00 (Sarajevo)",
		30: "UTC+01:00 (West Central Africa)",
		31: "UTC+02:00 (Athens/Istanbul)",
		32: "UTC+02:00 (Bucharest)",
		33: "UTC+02:00 (Cairo)",
		34: "UTC+02:00 (Harare/Pretoria)",
		35: "UTC+02:00 (Helsinki/Kyiv)",
		36: "UTC+02:00 (Jerusalem)",
		37: "UTC+03:00 (Baghdad)",
		38: "UTC+03:00 (Kuwait/Riyadh)",
		39: "UTC+03:00 (Moscow)",
		40: "UTC+03:00 (Nairobi)",
		41: "UTC+03:30 (Tehran)",
		42: "UTC+04:00 (Abu Dhabi/Muscat)",
		43: "UTC+04:00 (Baku/Tbilisi)",
		44: "UTC+04:30 (Kabul)",
		45: "UTC+05:00 (Islamabad/Karachi)",
		46: "UTC+05:00 (Ekaterinburg)",
		47: "UTC+05:30 (Mumbai/Kolkata)",
		48: "UTC+05:45 (Kathmandu)",
		49: "UTC+06:00 (Almaty/Dhaka)",
		50: "UTC+06:00 (Sri Jayawardenepura)",
		51: "UTC+06:30 (Rangoon)",
		52: "UTC+07:00 (Bangkok/Hanoi)",
		53: "UTC+07:00 (Krasnoyarsk)",
		54: "UTC+08:00 (Beijing/Hong Kong)",
		55: "UTC+08:00 (Kuala Lumpur/Singapore)",
		56: "UTC+08:00 (Taipei)",
		57: "UTC+08:00 (Perth)",
		58: "UTC+08:00 (Irkutsk)",
		59: "UTC+09:00 (Seoul)",
		60: "UTC+09:00 (Tokyo)",
		61: "UTC+09:00 (Yakutsk)",
		62: "UTC+09:30 (Darwin)",
		63: "UTC+09:30 (Adelaide)",
		64: "UTC+10:00 (Canberra/Sydney)",
		65: "UTC+10:00 (Brisbane)",
		66: "UTC+10:00 (Hobart)",
		67: "UTC+10:00 (Vladivostok)",
		68: "UTC+10:00 (Guam)",
		69: "UTC+11:00 (Magadan/Solomon Is.)",
		70: "UTC+12:00 (Auckland/Wellington)",
		71: "UTC+12:00 (Fiji)",
		72: "UTC+13:00 (Nuku'alofa)",
		73: "UTC+02:00 (Chisinau)",
		74: "UTC+02:00 (Amman)",
		75: "UTC+03:00 (Europe/Istanbul)",
		76: "UTC-04:30 (Caracas)",
		77: "UTC+08:00 (Ulaanbaatar)",
		78: "UTC+12:00 (Petropavlovsk-Kamchatsky)",
		79: "UTC+02:00 (Beirut)",
		80: "UTC+01:00 (Windhoek)",
		81: "UTC+04:00 (Yerevan)",
		82: "UTC+00:00 (UTC)",
		83: "UTC-03:00 (Montevideo)",
		84: "UTC-04:00 (Asuncion)",
		85: "UTC+02:00 (Kaliningrad)",
		86: "UTC+06:00 (Novosibirsk)",
		87: "UTC+11:00 (Srednekolymsk)",
	}
	if tz, ok := zones[idx]; ok {
		return tz
	}
	return fmt.Sprintf("Unknown (index %d)", idx)
}

// Top-level WatchGuard configuration
type WatchGuardConfig struct {
	XMLName            xml.Name          `xml:"profile" json:"-"`
	ForVersion         string            `xml:"for-version" json:"for_version"`
	SystemParameters   SystemParameters  `xml:"system-parameters" json:"system_parameters"`
	PolicyObjects      PolicyObjects     `xml:"policy-objects" json:"policy_objects"`
	AliasList          []Alias           `xml:"alias-list>alias" json:"-"`
	AddressGroupList   []AddressGroup    `xml:"address-group-list>address-group" json:"-"`
	PolicyList         PolicyList        `xml:"policy-list" json:"policy_list"`
	ServiceList        []ServiceDef      `xml:"service-list>service" json:"-"`
	NATList            []NATRule         `xml:"nat-list>nat" json:"-"`
	SecurityServices   SecurityServices  `xml:"security-services" json:"security_services"`
	ProxyActionList    ProxyActionList   `xml:"proxy-action-list" json:"proxy_action_list"`
	// VPN structures for Rule 8
	IKEActionList      []IKEAction       `xml:"ike-action-list>ike-action" json:"-"`
	IPsecProposalList  []IPsecProposal   `xml:"ipsec-proposal-list>ipsec-proposal" json:"-"`
	// Certificate list (alternative path) for Rule 9
	CertList           []IKECert         `xml:"cert-list>cert" json:"-"`
}

type SystemParameters struct {
	AdminUsers      []AdminUser        `xml:"admin-users>user"`
	DeviceConf      DeviceConf         `xml:"device-conf"`
	Interfaces      []Interface        `xml:"interface-list>interface"`
	IKECerts        []IKECert          `xml:"ike>ike-cert-list>cert"`
	DNSServers      []string           `xml:"dns-server-list>dns-entry"`
	LogConf         *LogConf           `xml:"log-conf"`
	CommonLogging   *CommonLogging     `xml:"common-logging"`
	DNSWatch        *DNSWatchConf      `xml:"dnswatch"`
	IPS             *GlobalIPS         `xml:"ips"`
	APT             *GlobalAPT         `xml:"apt"`
	BotnetDetection *GlobalBotnet      `xml:"botnet-detection"`
	DoSPrevention   []DoSItem          `xml:"dos-prevention>dos-item"`
	// Account lockout settings for Rule 7
	AuthGlobal      *AuthGlobalSetting `xml:"auth-global-setting"`
}

type LogConf struct {
	RemoteLogging *RemoteLogging `xml:"remote-logging"`
}

type RemoteLogging struct {
	Enabled string         `xml:"enabled"`
	Entries []LoggingEntry `xml:"remote-logging-list>logging-entry"`
}

type LoggingEntry struct {
	ServerIP   string `xml:"server-ip"`
	ServerPort string `xml:"server-port"`
}

type CommonLogging struct {
	Enabled   string     `xml:"enabled"`
	LogServer *LogServer `xml:"log-server"`
}

type LogServer struct {
	Type string `xml:"type"`
	Host string `xml:"host"`
	Port string `xml:"port"`
}

type DNSWatchConf struct {
	Enabled string `xml:"enabled"`
}

type DeviceConf struct {
	Model         string `xml:"for-model"`
	SystemName    string `xml:"system-name"`
	DomainName    string `xml:"domain-name"`
	SystemContact string `xml:"system-contact"`
	Location      string `xml:"location"`
	TimeZone      int    `xml:"time-zone"`
}

type Interface struct {
	Name    string   `xml:"name"`
	Items   []IfItem `xml:"if-item-list>item"`
}

type IfItem struct {
	ItemType   int        `xml:"item-type"`
	PhysicalIf *PhysicalIf `xml:"physical-if"`
	VlanIf     *VlanIf     `xml:"vlan-if"`
}

type PhysicalIf struct {
	IfDevName  string      `xml:"if-dev-name"`
	Enabled    int         `xml:"enabled"`
	IfProperty int         `xml:"if-property"`
	IP         string      `xml:"ip"`
	Netmask    string      `xml:"netmask"`
	LinkSpeed  int         `xml:"link-speed"`
	ExternalIf *ExternalIf `xml:"external-if"`
}

type ExternalIf struct {
	ExternalType int         `xml:"external-type"`
	DHCPClient   *DHCPClient `xml:"dhcp-client"`
}

type DHCPClient struct {
	HostName string `xml:"host-name"`
	ClientID string `xml:"client-id"`
}

type VlanIf struct {
	IfDevName  string `xml:"if-dev-name"`
	VlanID     int    `xml:"vlan-id"`
	VifProperty int   `xml:"vif-property"`
	IP         string `xml:"ip"`
	Netmask    string `xml:"netmask"`
}

type IKECert struct {
	Issuer  string `xml:"issuer"`
	Subject string `xml:"subject"`
}

// ── Account lockout settings ────────────────────────────────────────────────

type AuthGlobalSetting struct {
	MgmtAcctLockout *MgmtAcctLockout `xml:"mgmt-acct-lockout"`
}

type MgmtAcctLockout struct {
	Enabled  string `xml:"enabled"`
	Failures string `xml:"failures"`
	Lockouts string `xml:"lockouts"`
	Duration string `xml:"duration"`
}

// ── VPN IKE / IPsec structures ──────────────────────────────────────────────

// IKEAction represents a Phase 1 IKE action with transform proposals.
type IKEAction struct {
	Name      string       `xml:"name"`
	Transform IKETransform `xml:"ike-transform"`
}

// IKETransform holds the transform set members for Phase 1.
type IKETransform struct {
	Members []IKETransformMember `xml:"member"`
}

// IKETransformMember holds the cryptographic parameters for a single Phase 1 proposal.
type IKETransformMember struct {
	DHGroup    int `xml:"dh-group"`
	EncrypAlgm int `xml:"encryp-algm"`
	AuthAlgm   int `xml:"auth-algm"`
}

// IPsecProposal represents a Phase 2 IPsec transform set.
type IPsecProposal struct {
	Name         string       `xml:"name"`
	ESPTransform ESPTransform `xml:"esp-transform"`
}

// ESPTransform holds the ESP transform set members for Phase 2.
type ESPTransform struct {
	Members []ESPTransformMember `xml:"member"`
}

// ESPTransformMember holds the cryptographic parameters for a single Phase 2 proposal.
type ESPTransformMember struct {
	EncrypAlgm int `xml:"encryp-algm"`
	AuthAlgm   int `xml:"auth-algm"`
}

// DeviceInfo is the extracted device summary for the frontend
type DeviceInfo struct {
	Model         string          `json:"model"`
	SerialNumber  string          `json:"serial_number"`
	FirmwareVer   string          `json:"firmware_version"`
	SystemName    string          `json:"system_name"`
	DomainName    string          `json:"domain_name"`
	Contact       string          `json:"contact"`
	Location      string          `json:"location"`
	TimeZone      string          `json:"time_zone"`
	DNSServers    []string        `json:"dns_servers"`
	Interfaces    []InterfaceInfo `json:"interfaces"`
	LogServer     string          `json:"log_server"`
	SyslogServer  string          `json:"syslog_server,omitempty"`
	DNSWatch      string          `json:"dnswatch"`
}

type InterfaceInfo struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Device    string `json:"device"`
	IP        string `json:"ip"`
	Netmask   string `json:"netmask"`
	LinkSpeed int    `json:"link_speed,omitempty"`
	Enabled   bool   `json:"enabled"`
}

func ExtractDeviceInfo(cfg *WatchGuardConfig) DeviceInfo {
	info := DeviceInfo{
		Model:        cfg.SystemParameters.DeviceConf.Model,
		FirmwareVer:  cfg.ForVersion,
		SystemName:   cfg.SystemParameters.DeviceConf.SystemName,
		DomainName:   cfg.SystemParameters.DeviceConf.DomainName,
		Contact:      cfg.SystemParameters.DeviceConf.SystemContact,
		Location:     cfg.SystemParameters.DeviceConf.Location,
		TimeZone:     wgTimeZone(cfg.SystemParameters.DeviceConf.TimeZone),
		DNSServers:   cfg.SystemParameters.DNSServers,
	}

	// Extract serial number from IKE cert issuer
	for _, cert := range cfg.SystemParameters.IKECerts {
		if m := snRe.FindStringSubmatch(cert.Issuer); len(m) > 1 {
			info.SerialNumber = m[1]
			break
		}
	}

	// Log server (WatchGuard Dimension / WSM)
	if cl := cfg.SystemParameters.CommonLogging; cl != nil {
		if cl.LogServer != nil && cl.LogServer.Host != "" {
			info.LogServer = cl.LogServer.Host + ":" + cl.LogServer.Port
		}
		if cl.Enabled != "1" && cl.Enabled != "true" {
			info.LogServer = "Disabled"
		}
	} else {
		info.LogServer = "Not Configured"
	}

	// Syslog (remote logging)
	if lc := cfg.SystemParameters.LogConf; lc != nil && lc.RemoteLogging != nil {
		rl := lc.RemoteLogging
		if rl.Enabled == "1" || rl.Enabled == "true" {
			if len(rl.Entries) > 0 {
				e := rl.Entries[0]
				info.SyslogServer = e.ServerIP + ":" + e.ServerPort
			}
		}
	}

	// DNSWatch
	if dw := cfg.SystemParameters.DNSWatch; dw != nil {
		if dw.Enabled == "1" || dw.Enabled == "true" {
			info.DNSWatch = "Enabled"
		} else {
			info.DNSWatch = "Disabled"
		}
	} else {
		info.DNSWatch = "Not Configured"
	}

	// Extract interface info
	for _, iface := range cfg.SystemParameters.Interfaces {
		if strings.HasPrefix(iface.Name, "SSL-VPN") || strings.HasPrefix(iface.Name, "Azure") {
			continue
		}
		for _, item := range iface.Items {
			if item.PhysicalIf != nil {
				pif := item.PhysicalIf
				ifType := ifPropertyToType(pif.IfProperty)
				ip := pif.IP
				if pif.ExternalIf != nil && pif.ExternalIf.ExternalType == 2 {
					ip = "DHCP"
				}
				info.Interfaces = append(info.Interfaces, InterfaceInfo{
					Name:      iface.Name,
					Type:      ifType,
					Device:    pif.IfDevName,
					IP:        ip,
					Netmask:   pif.Netmask,
					LinkSpeed: pif.LinkSpeed,
					Enabled:   pif.Enabled == 1,
				})
			} else if item.VlanIf != nil {
				vif := item.VlanIf
				ifType := vifPropertyToType(vif.VifProperty)
				info.Interfaces = append(info.Interfaces, InterfaceInfo{
					Name:    iface.Name,
					Type:    ifType,
					Device:  vif.IfDevName,
					IP:      vif.IP,
					Netmask: vif.Netmask,
					Enabled: true,
				})
			}
		}
	}

	return info
}

func ifPropertyToType(p int) string {
	switch p {
	case 0:
		return "Mixed"
	case 1:
		return "Trusted"
	case 2:
		return "External"
	case 5:
		return "Optional"
	default:
		return "Other"
	}
}

func vifPropertyToType(p int) string {
	switch p {
	case 0:
		return "External"
	case 1:
		return "Trusted"
	case 2:
		return "Optional"
	default:
		return "Other"
	}
}

type AdminUser struct {
	Name     string `xml:"name,attr"`
	Password string `xml:"password"`
	Role     string `xml:"role"`
}

type PolicyObjects struct {
	Aliases []Alias `xml:"alias-list>alias"`
}

type Alias struct {
	NameAttr    string        `xml:"name,attr" json:"-"`
	NameElement string        `xml:"name" json:"-"`
	RawMembers  []AliasMember `xml:"alias-member-list>alias-member" json:"-"`
	// Legacy format: <member>value</member>
	LegacyMembers []string `xml:"member" json:"-"`
	// Members is populated after parsing
	Members []string `json:"members" xml:"-"`
}

type AliasMember struct {
	Type      string `xml:"type"`
	User      string `xml:"user"`
	Address   string `xml:"address"`
	Interface string `xml:"interface"`
	AliasName string `xml:"alias-name"`
}

type AddressGroup struct {
	Name    string               `xml:"name"`
	Members []AddressGroupMember `xml:"addr-group-member>member"`
}

type AddressGroupMember struct {
	Type         string `xml:"type"`
	HostIPAddr   string `xml:"host-ip-addr"`
	IPNetAddr    string `xml:"ip-network-addr"`
	IPMask       string `xml:"ip-mask"`
	StartIPAddr  string `xml:"start-ip-addr"`
	EndIPAddr    string `xml:"end-ip-addr"`
	Domain       string `xml:"domain"`
	DynamicAddrs string `xml:"dynamic-addrs"`
}

// resolveAddressGroupMember returns a human-readable string for an address group member.
func resolveAddressGroupMember(m AddressGroupMember) string {
	switch m.Type {
	case "1": // Host IP
		return m.HostIPAddr
	case "2": // Subnet
		if m.IPNetAddr != "" && m.IPMask != "" {
			return m.IPNetAddr + "/" + m.IPMask
		}
	case "3": // IP Range
		if m.StartIPAddr != "" && m.EndIPAddr != "" {
			return m.StartIPAddr + "-" + m.EndIPAddr
		}
	case "8": // FQDN/Domain
		if m.Domain != "" {
			return "FQDN:" + m.Domain
		}
	case "10": // Dynamic address list
		if m.DynamicAddrs != "" {
			return "Dynamic:" + m.DynamicAddrs
		}
	}
	// Fallback: return whatever non-empty field we have
	if m.HostIPAddr != "" {
		return m.HostIPAddr
	}
	return ""
}

// ── Service definitions ─────────────────────────────────────────────────────

type ServiceDef struct {
	Name        string          `xml:"name" json:"name"`
	Description string          `xml:"description" json:"description,omitempty"`
	Property    int             `xml:"property" json:"property"`
	ProxyType   string          `xml:"proxy-type" json:"proxy_type,omitempty"`
	Items       []ServiceMember `xml:"service-item>member" json:"items"`
	IdleTimeout int             `xml:"idle-timeout" json:"idle_timeout"`
}

type ServiceMember struct {
	Type       int `xml:"type" json:"type"`
	Protocol   int `xml:"protocol" json:"protocol"`
	ServerPort int `xml:"server-port" json:"server_port"`
}

// ProtocolName returns the human-readable protocol name for a protocol number.
func ProtocolName(proto int) string {
	switch proto {
	case 0:
		return "Any"
	case 1:
		return "ICMP"
	case 6:
		return "TCP"
	case 17:
		return "UDP"
	case 47:
		return "GRE"
	case 50:
		return "ESP"
	case 51:
		return "AH"
	case 58:
		return "ICMPv6"
	default:
		return fmt.Sprintf("IP/%d", proto)
	}
}

// ServicePortSummary returns a compact string like "TCP/443, UDP/53".
func ServicePortSummary(svc ServiceDef) string {
	if len(svc.Items) == 0 {
		return "any"
	}
	parts := make([]string, 0, len(svc.Items))
	for _, m := range svc.Items {
		proto := ProtocolName(m.Protocol)
		if m.ServerPort == 0 && m.Protocol == 0 {
			return "any"
		}
		if m.ServerPort == 0 {
			parts = append(parts, proto)
		} else {
			parts = append(parts, fmt.Sprintf("%s/%d", proto, m.ServerPort))
		}
	}
	return strings.Join(parts, ", ")
}

// ── NAT rules ───────────────────────────────────────────────────────────────

type NATRule struct {
	Name        string      `xml:"name" json:"name"`
	Description string      `xml:"description" json:"description,omitempty"`
	Property    int         `xml:"property" json:"property"`
	Type        int         `xml:"type" json:"type"`
	Algorithm   int         `xml:"algorithm" json:"algorithm"`
	ProxyARP    int         `xml:"proxy-arp" json:"proxy_arp"`
	Items       []NATMember `xml:"nat-item>member" json:"items,omitempty"`
}

type NATMember struct {
	AddrType    int    `xml:"addr-type" json:"addr_type"`
	Port        int    `xml:"port" json:"port"`
	IP          string `xml:"ip" json:"ip,omitempty"`
	ExtAddrName string `xml:"ext-addr-name" json:"ext_addr_name,omitempty"`
	Interface   string `xml:"interface" json:"interface,omitempty"`
	AddrName    string `xml:"addr-name" json:"addr_name,omitempty"`
}

// NATTypeName returns a human-readable NAT type.
func NATTypeName(t int) string {
	switch t {
	case 3:
		return "Dynamic NAT"
	case 4:
		return "DNAT"
	case 7:
		return "SNAT"
	default:
		return fmt.Sprintf("NAT-Type-%d", t)
	}
}

// ── Global security settings ────────────────────────────────────────────────

type GlobalIPS struct {
	Enabled      string          `xml:"enabled" json:"enabled"`
	FullScanMode string          `xml:"full-scan-mode" json:"full_scan_mode"`
	ThreatLevel  *IPSThreatLevel `xml:"threat-level" json:"threat_level,omitempty"`
}

type IPSThreatLevel struct {
	Critical IPSThreatAction `xml:"critical" json:"critical"`
	High     IPSThreatAction `xml:"high" json:"high"`
	Medium   IPSThreatAction `xml:"medium" json:"medium"`
	Low      IPSThreatAction `xml:"low" json:"low"`
	Info     IPSThreatAction `xml:"info" json:"info"`
}

type IPSThreatAction struct {
	Action string `xml:"action" json:"action"`
	Alarm  string `xml:"alarm" json:"alarm"`
	Log    string `xml:"log" json:"log"`
}

type GlobalAPT struct {
	Enabled string `xml:"enabled" json:"enabled"`
}

type GlobalBotnet struct {
	Enabled string `xml:"enabled" json:"enabled"`
}

type DoSItem struct {
	Type      int `xml:"dos-type" json:"type"`
	Enabled   int `xml:"dos-enable" json:"enabled"`
	Threshold int `xml:"dos-threshold" json:"threshold"`
}

// Name returns the alias name, preferring the attribute form over the element form.
func (a Alias) Name() string {
	if a.NameAttr != "" {
		return a.NameAttr
	}
	return a.NameElement
}

// MarshalJSON implements custom JSON marshaling to output "name" field.
func (a Alias) MarshalJSON() ([]byte, error) {
	type plain struct {
		Name    string   `json:"name"`
		Members []string `json:"members"`
	}
	return json.Marshal(plain{Name: a.Name(), Members: a.Members})
}

// resolveAliasMembers populates the Members field from raw parsed data.
// addrGroups maps address-group names to their resolved IP addresses.
func resolveAliasMembers(a *Alias, addrGroups map[string]AddressGroup) {
	if len(a.LegacyMembers) > 0 {
		a.Members = a.LegacyMembers
		return
	}
	for _, m := range a.RawMembers {
		switch m.Type {
		case "2": // alias reference
			if m.AliasName != "" {
				a.Members = append(a.Members, m.AliasName)
			}
		case "1": // user/address/interface
			if m.User != "" && m.User != "Any" {
				a.Members = append(a.Members, m.User)
			}
			if m.Address != "" && m.Address != "Any" {
				// Try to resolve address-group reference to IP
				if ag, ok := addrGroups[m.Address]; ok {
					for _, gm := range ag.Members {
						if s := resolveAddressGroupMember(gm); s != "" {
							a.Members = append(a.Members, s)
						}
					}
				} else {
					a.Members = append(a.Members, m.Address)
				}
			}
		default:
			if m.AliasName != "" {
				a.Members = append(a.Members, m.AliasName)
			}
		}
	}
}

type PolicyList struct {
	Policies []Policy `xml:"policy"`
}

// ResolvedNAT holds the enriched NAT detail for a policy.
type ResolvedNAT struct {
	Name     string   `json:"name"`
	TypeName string   `json:"type_name"`
	Members  []string `json:"members,omitempty"`
}

type PolicyProxyServices struct {
	GatewayAV       bool   `json:"gateway_av"`
	WebBlocker      bool   `json:"web_blocker"`
	WebBlockerProfile string `json:"web_blocker_profile,omitempty"`
	APTBlocker      bool   `json:"apt_blocker"`
	IPS             bool   `json:"ips"`
}

type Policy struct {
	Order         int                  `xml:"-" json:"order"`
	Name          string               `xml:"name" json:"name"`
	PolicyType    string               `xml:"type" json:"type"`
	Enabled       string               `xml:"enable" json:"enabled"`
	Description   string               `xml:"description" json:"description"`
	Schedule      string               `xml:"schedule" json:"schedule,omitempty"`
	From          PolicyFrom           `xml:"from-alias-list" json:"from"`
	To            PolicyTo             `xml:"to-alias-list" json:"to"`
	Service       string               `xml:"service" json:"service"`
	Action        string               `xml:"firewall-action" json:"action"`
	// Raw log fields parsed from XML sibling elements
	LogRaw        string               `xml:"log" json:"-"`
	LogForReport  string               `xml:"log-for-report" json:"-"`
	// Logging is populated programmatically after parsing
	Logging       LogSettings          `xml:"-" json:"logging"`
	Proxy         string               `xml:"proxy" json:"proxy"`
	IPSMonitor    string               `xml:"ips-monitor-enabled" json:"ips_monitor_enabled"`
	AppAction     string               `xml:"app-action" json:"app_action"`
	NATRef        string               `xml:"nat" json:"nat_ref,omitempty"`
	ServicePorts  string               `xml:"-" json:"service_ports,omitempty"`
	ResolvedNAT   *ResolvedNAT         `xml:"-" json:"resolved_nat,omitempty"`
	ProxyServices *PolicyProxyServices `xml:"-" json:"proxy_services,omitempty"`
	IsSystem      bool                 `xml:"-" json:"is_system"`
}

type PolicyFrom struct {
	Aliases []string `xml:"alias" json:"aliases"`
}

type PolicyTo struct {
	Aliases []string `xml:"alias" json:"aliases"`
}

// LogSettings holds policy logging state for JSON output.
// Populated programmatically from raw XML fields, not directly via xml tags.
type LogSettings struct {
	Enabled   string `json:"enabled"`
	ForReport string `json:"for_report"`
}

type ProxyActionList struct {
	ProxyActions []ProxyAction `xml:"proxy-action"`
}

type ProxyAction struct {
	Name         string `xml:"proxy-name"`
	Description  string `xml:"proxy-description"`
	Type         string `xml:"proxy-type-attr"`
	LogForReport string `xml:"log-for-report"`
	HTTP         *HTTPProxyAction  `xml:"http"`
	HTTPS       *HTTPSProxyAction `xml:"https"`
	TCP         *TCPProxyAction   `xml:"outgoing"`
}

type HTTPProxyAction struct {
	// Gateway AntiVirus: <anti-virus><enabled>true</enabled>
	GatewayAV  string `xml:"anti-virus>enabled"`
	// APT Blocker: <apt-enabled>true</apt-enabled>
	APTBlocker string `xml:"apt-enabled"`
	// WebBlocker: lives under request>uri>filter>helper-name — non-empty means configured
	WebBlocker string `xml:"request>uri>filter>helper-name"`
}

type HTTPSProxyAction struct {
	// HTTPS proxy redirects content inspection to an HTTP proxy action
	RedirectTo string `xml:"redirect-to"`
	// WebBlocker via Deep Packet Inspection: <wb-inspect websense="true">
	WebBlockerInspect string `xml:"wb-inspect>websense"`
}

type TCPProxyAction struct {
	Redirects []TCPRedirectRule `xml:"protocols>rule"`
}

// TCPRedirectRule maps a protocol pattern (http, ssl, ftp…) to a sub-proxy action name.
type TCPRedirectRule struct {
	Name    string `xml:"name"`    // e.g. "HTTP-Client.Arda"
	Pattern string `xml:"pattern"` // e.g. "http", "ssl"
}

type Service struct {
	Name    string `xml:"name,attr"`
	Enabled string `xml:"enabled,attr"`
}

type SecurityServices struct {
	GatewayAV  *ServiceGlobal `xml:"gateway-av"`
	IPS        *ServiceGlobal `xml:"intrusion-prevention"`
	WebBlocker *ServiceGlobal `xml:"webblocker"`
	APTBlocker *ServiceGlobal `xml:"apt-blocker"`
}

type ServiceGlobal struct {
	Enabled string `xml:"enabled,attr"`
}

func ParseConfig(data []byte) (*WatchGuardConfig, error) {
	var config WatchGuardConfig
	if err := xml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
