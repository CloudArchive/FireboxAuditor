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

func FetchConfigViaSSH(cfg SSHConfig) ([]byte, []string, error) {
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
		return nil, logs, fmt.Errorf("SSH bağlantısı kurulamadı (%s): %w", addr, err)
	}
	defer client.Close()
	log("SSH connection established")

	session, err := client.NewSession()
	if err != nil {
		log(fmt.Sprintf("Session error: %v", err))
		return nil, logs, fmt.Errorf("SSH oturumu açılamadı: %w", err)
	}
	defer session.Close()
	log("SSH session created")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	cmd := "export config to console"
	log(fmt.Sprintf("Running command: %s", cmd))
	if err := session.Run(cmd); err != nil {
		log(fmt.Sprintf("Command error: %v", err))
		if stderr.Len() > 0 {
			log(fmt.Sprintf("Stderr: %s", stderr.String()))
		}
		return nil, logs, fmt.Errorf("komut çalıştırılamadı: %w (stderr: %s)", err, stderr.String())
	}
	log("Command executed successfully")

	output := stdout.String()

	// Extract XML portion from output
	startIdx := strings.Index(output, "<?xml")
	if startIdx == -1 {
		startIdx = strings.Index(output, "<profile")
	}
	if startIdx == -1 {
		log("Error: XML configuration not found in output")
		return nil, logs, fmt.Errorf("XML konfigürasyonu çıktıda bulunamadı")
	}

	log("Config fetched and parsed successfully")
	return []byte(output[startIdx:]), logs, nil
}
