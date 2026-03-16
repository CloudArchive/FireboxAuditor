package main

import (
	"encoding/xml"
	"regexp"
	"strings"
)

// Top-level WatchGuard configuration
type WatchGuardConfig struct {
	XMLName          xml.Name          `xml:"profile"`
	ForVersion       string            `xml:"for-version"`
	SystemParameters SystemParameters  `xml:"system-parameters"`
	PolicyObjects    PolicyObjects     `xml:"policy-objects"`
	PolicyList       PolicyList        `xml:"policy-list"`
	SecurityServices SecurityServices  `xml:"security-services"`
}

type SystemParameters struct {
	AdminUsers  []AdminUser `xml:"admin-users>user"`
	DeviceConf  DeviceConf  `xml:"device-conf"`
	Interfaces  []Interface `xml:"interface-list>interface"`
	IKECerts    []IKECert   `xml:"ike>ike-cert-list>cert"`
	DNSServers  []string    `xml:"dns-server-list>dns-entry"`
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
	snRe := regexp.MustCompile(`SN\s+(\S+)`)
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
	Name    string   `xml:"name,attr"`
	Members []string `xml:"member"`
}

type PolicyList struct {
	Policies []Policy `xml:"policy"`
}

type Policy struct {
	Name        string       `xml:"name,attr"`
	PolicyType  string       `xml:"type,attr"`
	Enabled     string       `xml:"enabled,attr"`
	From        PolicyFrom   `xml:"from"`
	To          PolicyTo     `xml:"to"`
	Service     string       `xml:"service,attr"`
	Logging     LogSettings  `xml:"log"`
	ProxyAction *ProxyAction `xml:"proxy-action"`
}

type PolicyFrom struct {
	Aliases []string `xml:"alias"`
}

type PolicyTo struct {
	Aliases []string `xml:"alias"`
}

type LogSettings struct {
	Enabled    string `xml:"enabled,attr"`
	ForReport  string `xml:"for-report,attr"`
	LogMessage string `xml:"log-message,attr"`
}

type ProxyAction struct {
	Name     string    `xml:"name,attr"`
	Services []Service `xml:"service"`
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
