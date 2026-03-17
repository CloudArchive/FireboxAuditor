package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

//go:embed static/*
var staticFS embed.FS

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8443"
	}

	// Initialize auth (creates default user if needed)
	initAuth()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// API routes
	setupRoutes(r)

	// Serve embedded React frontend
	distFS, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatal("Static dosyalar yüklenemedi:", err)
	}
	fileServer := http.FileServer(http.FS(distFS))

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		f, err := distFS.Open(path[1:])
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}
		// SPA fallback
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	log.Printf("Firebox Auditor başlatılıyor: http://localhost:%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Sunucu başlatılamadı:", err)
	}
}
