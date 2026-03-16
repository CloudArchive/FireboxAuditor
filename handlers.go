package main

import (
	"io"
	"net/http"

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
	c.JSON(http.StatusOK, report)
}

func handleSSH(c *gin.Context) {
	var sshCfg SSHConfig
	if err := c.ShouldBindJSON(&sshCfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz bağlantı bilgileri: " + err.Error()})
		return
	}

	data, err := FetchConfigViaSSH(sshCfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
