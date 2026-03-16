package main

import "encoding/xml"

// Top-level WatchGuard configuration
type WatchGuardConfig struct {
	XMLName          xml.Name          `xml:"profile"`
	SystemParameters SystemParameters  `xml:"system-parameters"`
	PolicyObjects    PolicyObjects     `xml:"policy-objects"`
	PolicyList       PolicyList        `xml:"policy-list"`
	SecurityServices SecurityServices  `xml:"security-services"`
}

type SystemParameters struct {
	AdminUsers []AdminUser `xml:"admin-users>user"`
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
