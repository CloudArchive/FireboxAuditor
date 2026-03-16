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

func FetchConfigViaSSH(cfg SSHConfig) ([]byte, error) {
	if cfg.Port == 0 {
		cfg.Port = 4118
	}

	sshConfig := &ssh.ClientConfig{
		User: cfg.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(cfg.Password),
			ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
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
		return nil, fmt.Errorf("SSH bağlantısı kurulamadı (%s): %w", addr, err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("SSH oturumu açılamadı: %w", err)
	}
	defer session.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run("export config to console"); err != nil {
		return nil, fmt.Errorf("komut çalıştırılamadı: %w (stderr: %s)", err, stderr.String())
	}

	output := stdout.String()

	// Extract XML portion from output
	startIdx := strings.Index(output, "<?xml")
	if startIdx == -1 {
		startIdx = strings.Index(output, "<profile")
	}
	if startIdx == -1 {
		return nil, fmt.Errorf("XML konfigürasyonu çıktıda bulunamadı")
	}

	return []byte(output[startIdx:]), nil
}
