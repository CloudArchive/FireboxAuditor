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

// Top-level WatchGuard configuration
type WatchGuardConfig struct {
	XMLName           xml.Name          `xml:"profile" json:"-"`
	ForVersion        string            `xml:"for-version" json:"for_version"`
	SystemParameters  SystemParameters  `xml:"system-parameters" json:"system_parameters"`
	PolicyObjects     PolicyObjects     `xml:"policy-objects" json:"policy_objects"`
	AliasList         []Alias           `xml:"alias-list>alias" json:"-"`
	AddressGroupList  []AddressGroup    `xml:"address-group-list>address-group" json:"-"`
	PolicyList        PolicyList        `xml:"policy-list" json:"policy_list"`
	ServiceList       []ServiceDef      `xml:"service-list>service" json:"-"`
	NATList           []NATRule         `xml:"nat-list>nat" json:"-"`
	SecurityServices  SecurityServices  `xml:"security-services" json:"security_services"`
	ProxyActionList   ProxyActionList   `xml:"proxy-action-list" json:"proxy_action_list"`
}

type SystemParameters struct {
	AdminUsers      []AdminUser      `xml:"admin-users>user"`
	DeviceConf      DeviceConf       `xml:"device-conf"`
	Interfaces      []Interface      `xml:"interface-list>interface"`
	IKECerts        []IKECert        `xml:"ike>ike-cert-list>cert"`
	DNSServers      []string         `xml:"dns-server-list>dns-entry"`
	IPS             *GlobalIPS       `xml:"ips"`
	APT             *GlobalAPT       `xml:"apt"`
	BotnetDetection *GlobalBotnet    `xml:"botnet-detection"`
	DoSPrevention   []DoSItem        `xml:"dos-prevention>dos-item"`
}

type DeviceConf struct {
	Model         string `xml:"for-model"`
	SystemName    string `xml:"system-name"`
	DomainName    string `xml:"domain-name"`
	SystemContact string `xml:"system-contact"`
	Location      string `xml:"location"`
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
	Issuer string `xml:"issuer"`
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
	DNSServers    []string        `json:"dns_servers"`
	Interfaces    []InterfaceInfo `json:"interfaces"`
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
		DNSServers:   cfg.SystemParameters.DNSServers,
	}

	// Extract serial number from IKE cert issuer
	for _, cert := range cfg.SystemParameters.IKECerts {
		if m := snRe.FindStringSubmatch(cert.Issuer); len(m) > 1 {
			info.SerialNumber = m[1]
			break
		}
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

type NATSettings struct {
	Dynamic string `xml:"dynamic" json:"dynamic,omitempty"`
	Static  string `xml:"static" json:"static,omitempty"`
}

// Resolved NAT detail attached to policy after enrichment
type ResolvedNAT struct {
	Dynamic *NATRuleSummary `json:"dynamic,omitempty"`
	Static  *NATRuleSummary `json:"static,omitempty"`
}

type NATRuleSummary struct {
	Name     string `json:"name"`
	TypeName string `json:"type_name"`
	IP       string `json:"ip,omitempty"`
	AddrName string `json:"addr_name,omitempty"`
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
	// Raw log fields parsed from XML sibling elements
	LogRaw        string               `xml:"log" json:"-"`
	LogForReport  string               `xml:"log-for-report" json:"-"`
	// Logging is populated programmatically after parsing
	Logging       LogSettings          `xml:"-" json:"logging"`
	Proxy         string               `xml:"proxy" json:"proxy"`
	IPSMonitor    string               `xml:"ips-monitor-enabled" json:"ips_monitor_enabled"`
	AppAction     string               `xml:"app-action" json:"app_action"`
	NAT           *NATSettings         `xml:"nat" json:"nat,omitempty"`
	// Enriched fields populated after parsing
	ServicePorts  string               `xml:"-" json:"service_ports,omitempty"`
	ResolvedNAT   *ResolvedNAT         `xml:"-" json:"resolved_nat,omitempty"`
	ProxyServices *PolicyProxyServices `xml:"-" json:"proxy_services,omitempty"`
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
	Name        string `xml:"proxy-name"`
	Description string `xml:"proxy-description"`
	Type        string `xml:"proxy-type-attr"`
	HTTP        *HTTPProxyAction  `xml:"http"`
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
