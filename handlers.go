package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func setupRoutes(r *gin.Engine) {
	// Public
	r.POST("/api/auth/login", handleLogin)

	// Protected
	api := r.Group("/api", AuthMiddleware())
	{
		// History
		api.GET("/history", handleListHistory)
		api.GET("/history/:id", handleGetHistory)
		api.DELETE("/history/:id", handleDeleteHistory)

		// Audit
		api.POST("/audit/upload", handleUpload)

		// SSH enrichment (separate from audit)
		api.POST("/ssh/enrich", handleSSHEnrich)
	}
}

// ── History Handlers ──────────────────────────────────────────────────────────

func handleListHistory(c *gin.Context) {
	username := c.GetString("username")
	records, err := ListAudits(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Geçmiş yüklenemedi: " + err.Error()})
		return
	}
	// Return lightweight summaries (no full report payload)
	type Summary struct {
		ID         string    `json:"id"`
		CreatedAt  time.Time `json:"created_at"`
		FileName   string    `json:"file_name"`
		DeviceName string    `json:"device_name"`
		Score      int       `json:"score"`
		Enriched   bool      `json:"enriched"`
	}
	summaries := make([]Summary, 0, len(records))
	for _, r := range records {
		summaries = append(summaries, Summary{
			ID:         r.ID,
			CreatedAt:  r.CreatedAt,
			FileName:   r.FileName,
			DeviceName: r.DeviceName,
			Score:      r.Score,
			Enriched:   r.Enrichment != nil,
		})
	}
	c.JSON(http.StatusOK, summaries)
}

func handleGetHistory(c *gin.Context) {
	username := c.GetString("username")
	id := c.Param("id")

	record, err := GetAudit(username, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if record == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kayıt bulunamadı"})
		return
	}
	c.JSON(http.StatusOK, record)
}

func handleDeleteHistory(c *gin.Context) {
	username := c.GetString("username")
	id := c.Param("id")

	if err := DeleteAudit(username, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ── Upload + Audit ────────────────────────────────────────────────────────────

func handleUpload(c *gin.Context) {
	username := c.GetString("username")

	// File size limit: 10 MB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20)

	file, header, err := c.Request.FormFile("config")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "XML dosyası yüklenemedi: " + err.Error()})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Dosya okunamadı: " + err.Error()})
		return
	}

	cfg, err := ParseConfig(data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "XML parse hatası: " + err.Error()})
		return
	}

	report := RunAudit(cfg)

	// Build record
	deviceName := cfg.SystemParameters.DeviceConf.SystemName
	if deviceName == "" {
		deviceName = cfg.SystemParameters.DeviceConf.Model
	}
	if deviceName == "" {
		deviceName = "Firebox"
	}

	record := &AuditRecord{
		ID:         fmt.Sprintf("%d", time.Now().UnixMilli()),
		CreatedAt:  time.Now(),
		FileName:   header.Filename,
		DeviceName: deviceName,
		Score:      report.Score,
		Report:     report,
	}

	if err := SaveAudit(username, record); err != nil {
		// Non-fatal: return report even if save fails
		c.JSON(http.StatusOK, gin.H{
			"id":     record.ID,
			"report": report,
			"warn":   "Kayıt kaydedilemedi: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":     record.ID,
		"report": report,
	})
}

// ── SSH Enrich ────────────────────────────────────────────────────────────────

func handleSSHEnrich(c *gin.Context) {
	username := c.GetString("username")

	var req struct {
		AuditID string `json:"audit_id" binding:"required"`
		SSHConfig
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek: " + err.Error()})
		return
	}

	// Verify the audit belongs to this user
	record, err := GetAudit(username, req.AuditID)
	if err != nil || record == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Audit kaydı bulunamadı"})
		return
	}

	enrich, logs, err := EnrichFromSSH(req.SSHConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"logs":  logs,
		})
		return
	}

	// Persist enrichment
	if saveErr := UpdateEnrichment(username, req.AuditID, enrich); saveErr != nil {
		// Non-fatal
		c.JSON(http.StatusOK, gin.H{
			"enrichment": enrich,
			"logs":       logs,
			"warn":       "Zenginleştirme kaydedilemedi: " + saveErr.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enrichment": enrich,
		"logs":       logs,
	})
}

// ── SSH Enrich Logic (moved from ssh_client.go) ───────────────────────────────

func EnrichFromSSH(cfg SSHConfig) (*EnrichData, []string, error) {
	// Run sysinfo
	sysOutput, logs, err := ExecuteSSHCommand(cfg, "show sysinfo")
	if err != nil {
		return nil, logs, fmt.Errorf("sysinfo komutu başarısız: %w", err)
	}
	// Log raw output for debugging (truncated to 1000 chars)
	rawSys := stripANSI(sysOutput)
	if len(rawSys) > 1000 {
		rawSys = rawSys[:1000] + "..."
	}
	logs = append(logs, fmt.Sprintf("[RAW sysinfo]\n%s", rawSys))

	sysInfo := ParseSysInfo(stripSSHNoise(sysOutput))
	logs = append(logs, fmt.Sprintf("[PARSED] serial=%q uptime=%q cpu=%q mem=%q",
		sysInfo.SerialNumber, sysInfo.UpTime, sysInfo.CPUUsage, sysInfo.MemoryUsage))

	// Run show feature-key
	fkOutput, fkLogs, err := ExecuteSSHCommand(cfg, "show feature-key")
	logs = append(logs, fkLogs...)
	if err != nil {
		// Feature key is optional — don't fail the whole enrichment
		logs = append(logs, "[WARN] feature-key alınamadı: "+err.Error())
	} else {
		rawFK := stripANSI(fkOutput)
		if len(rawFK) > 500 {
			rawFK = rawFK[:500] + "..."
		}
		logs = append(logs, fmt.Sprintf("[RAW feature-key]\n%s", rawFK))
	}

	enrich := &EnrichData{
		SerialNumber: sysInfo.SerialNumber,
		FullVersion:  sysInfo.Version,
		UpTime:       sysInfo.UpTime,
		MemoryUsage:  sysInfo.MemoryUsage,
		CPUUsage:     sysInfo.CPUUsage,
		EnrichedAt:   time.Now(),
		SSHHost:      cfg.Host,
		SSHPort:      cfg.Port,
		SSHUsername:  cfg.Username,
	}

	if fkOutput != "" {
		// Strip ANSI/control chars and leading prompt lines
		clean := stripSSHNoise(stripANSI(fkOutput))
		enrich.FeatureKey = ParseFeatureKey(clean)
	}

	// Run show features
	fOutput, fLogs, err := ExecuteSSHCommand(cfg, "show features")
	logs = append(logs, fLogs...)
	if err != nil {
		logs = append(logs, "[WARN] features alınamadı: "+err.Error())
	} else {
		rawF := stripANSI(fOutput)
		if len(rawF) > 500 {
			rawF = rawF[:500] + "..."
		}
		logs = append(logs, fmt.Sprintf("[RAW features]\n%s", rawF))
		cleanF := stripSSHNoise(stripANSI(fOutput))
		enrich.Features = ParseShowFeatures(cleanF)
	}

	// Fallback: pull serial from feature-key output if sysinfo didn't have it
	if enrich.SerialNumber == "" && enrich.FeatureKey != nil {
		for _, line := range splitLines(enrich.FeatureKey.Raw) {
			t := trimSpace(line)
			if hasPrefix(t, "Serial Number:") {
				enrich.SerialNumber = trimSpace(t[len("Serial Number:"):])
				break
			}
		}
	}

	return enrich, logs, nil
}

// stripSSHNoise removes prompt lines and ANSI codes from CLI output.
func stripSSHNoise(s string) string {
	var clean []string
	for _, line := range splitLines(s) {
		// Skip prompt lines (WG# or WG>)
		stripped := trimSpace(line)
		if strings.HasPrefix(stripped, "WG#") || strings.HasPrefix(stripped, "WG>") {
			continue
		}
		clean = append(clean, line)
	}
	return strings.Join(clean, "\n")
}
