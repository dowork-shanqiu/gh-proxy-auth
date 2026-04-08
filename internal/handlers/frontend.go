package handlers

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var frontendFS http.FileSystem

func SetFrontendFS(fsys fs.FS) {
	frontendFS = http.FS(fsys)
}

func IsProxyPath(path string) bool {
	return checkURL(path) || checkURL("https://"+path)
}

func ServeFrontend(c *gin.Context) {
	if frontendFS == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Frontend not available"})
		return
	}

	path := c.Request.URL.Path

	// Try to serve static files
	if path != "/" && !strings.HasSuffix(path, "/") {
		file, err := frontendFS.Open(path)
		if err == nil {
			file.Close()
			http.FileServer(frontendFS).ServeHTTP(c.Writer, c.Request)
			return
		}
	}

	// For SPA routing, serve index.html
	c.Request.URL.Path = "/index.html"
	http.FileServer(frontendFS).ServeHTTP(c.Writer, c.Request)
}
