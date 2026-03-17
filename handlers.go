package main

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func setupRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.POST("/audit/upload", handleUpload)
		api.POST("/audit/ssh", handleSSH)
	}
}

func handleUpload(c *gin.Context) {
	file, _, err := c.Request.FormFile("config")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "XML dosyası yüklenemedi: " + err.Error()})
		return
	}
	defer file.Close()

	const maxUploadSize = 10 << 20 // 10 MB
	data, err := io.ReadAll(io.LimitReader(file, maxUploadSize))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Dosya okunamadı: " + err.Error()})
		return
	}
	if int64(len(data)) >= maxUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dosya boyutu 10 MB sınırını aşıyor"})
		return
	}

	cfg, err := ParseConfig(data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "XML parse hatası: " + err.Error()})
		return
	}

	report := RunAudit(cfg)
	c.JSON(http.StatusOK, report)
}

func handleSSH(c *gin.Context) {
	var req struct {
		SSHConfig
		Action string `json:"action"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek: " + err.Error()})
		return
	}

	command := ""
	switch req.Action {
	case "sysinfo":
		command = "sysinfo"
	case "audit":
		// Try both documented and undocumented export commands
		command = "export config to console"
	case "feature-key":
		command = "show feature-key"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz aksiyon"})
		return
	}

	output, logs, err := ExecuteSSHCommand(req.SSHConfig, command)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"logs":  logs,
		})
		return
	}

	if req.Action == "sysinfo" {
		sysInfo := ParseSysInfo(output)
		c.JSON(http.StatusOK, gin.H{
			"data": sysInfo,
			"logs": logs,
		})
		return
	}

	if req.Action == "feature-key" {
		c.JSON(http.StatusOK, gin.H{
			"data": output, // Return raw for now, can parse later if needed
			"logs": logs,
		})
		return
	}

	// Default to audit/config fetch
	startIdx := strings.Index(output, "<?xml")
	if startIdx == -1 {
		startIdx = strings.Index(output, "<profile")
	}

	if startIdx == -1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Konfigürasyon verisi bulunamadı. Cihazınız 'export config to console' komutunu desteklemiyor olabilir.",
			"logs":  logs,
		})
		return
	}

	xmlData := output[startIdx:]
	if endIdx := strings.Index(xmlData, "</profile>"); endIdx != -1 {
		xmlData = xmlData[:endIdx+len("</profile>")]
	}

	cfg, err := ParseConfig([]byte(xmlData))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "XML parse hatası: " + err.Error(),
			"logs":  logs,
		})
		return
	}

	report := RunAudit(cfg)
	c.JSON(http.StatusOK, gin.H{
		"report": report,
		"logs":   logs,
	})
}
