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
		// Try to serve the file; fall back to index.html for SPA routing
		path := c.Request.URL.Path
		f, err := distFS.Open(path[1:]) // strip leading "/"
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}
		// SPA fallback
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	tlsCert := os.Getenv("TLS_CERT")
	tlsKey := os.Getenv("TLS_KEY")

	if tlsCert != "" && tlsKey != "" {
		log.Printf("Firebox Auditor başlatılıyor (TLS): https://localhost:%s\n", port)
		if err := r.RunTLS(":"+port, tlsCert, tlsKey); err != nil {
			log.Fatal("Sunucu başlatılamadı:", err)
		}
	} else {
		log.Printf("Firebox Auditor başlatılıyor: http://localhost:%s\n", port)
		log.Println("UYARI: TLS aktif değil. TLS_CERT ve TLS_KEY ortam değişkenlerini ayarlayın.")
		if err := r.Run(":" + port); err != nil {
			log.Fatal("Sunucu başlatılamadı:", err)
		}
	}
}
