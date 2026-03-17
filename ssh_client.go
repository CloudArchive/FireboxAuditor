package main

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type SysInfo struct {
	SystemName    string `json:"system_name"`
	Model         string `json:"model"`
	Contact       string `json:"contact"`
	Location      string `json:"location"`
	SerialNumber  string `json:"serial_number"`
	Version       string `json:"version"`
	UpTime        string `json:"up_time"`
	MemoryUsage   string `json:"memory_usage"`
	CPUUsage      string `json:"cpu_usage"`
}

func ExecuteSSHCommand(cfg SSHConfig, command string) (string, []string, error) {
	var logs []string
	log := func(msg string) {
		logs = append(logs, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg))
	}

	if cfg.Port == 0 {
		cfg.Port = 4118
	}

	log(fmt.Sprintf("Connecting to %s:%d...", cfg.Host, cfg.Port))

	sshConfig := &ssh.ClientConfig{
		User: cfg.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(cfg.Password),
			ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
				log("Keyboard-interactive authentication requested")
				answers := make([]string, len(questions))
				for i := range questions {
					answers[i] = cfg.Password
				}
				return answers, nil
			}),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}

	addr := net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", cfg.Port))
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		log(fmt.Sprintf("Dial error: %v", err))
		return "", logs, fmt.Errorf("SSH bağlantısı kurulamadı: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", logs, fmt.Errorf("SSH oturumu açılamadı: %w", err)
	}
	defer session.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	log(fmt.Sprintf("Running command: %s", command))
	if err := session.Run(command); err != nil {
		return "", logs, fmt.Errorf("komut başarısız: %w (stderr: %s)", err, stderr.String())
	}

	return stdout.String(), logs, nil
}

func ParseSysInfo(output string) SysInfo {
	info := SysInfo{}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSpace(strings.ToLower(parts[0]))
		val := strings.TrimSpace(parts[1])

		switch key {
		case "system name":
			info.SystemName = val
		case "system model":
			info.Model = val
		case "contact":
			info.Contact = val
		case "location":
			info.Location = val
		case "serial number":
			info.SerialNumber = val
		case "version":
			info.Version = val
		case "up time":
			info.UpTime = val
		case "memory usage":
			info.MemoryUsage = val
		case "cpu utilization":
			info.CPUUsage = val
		}
	}
	return info
}
