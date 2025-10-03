package handlers

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

func PageContentHandler(c echo.Context) error {
	id := c.Param("id")
	pagePath := filepath.Join("web/content", id, "index.html")
	return c.File(pagePath)
}

func PageFullHandler(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		id = "home"
	}

	// read content
	pagePath := filepath.Join("web/content", id, "index.html")
	contentBytes, err := os.ReadFile(pagePath)
	if err != nil {
		if os.IsNotExist(err) {
			return c.String(http.StatusNotFound, "Page not found")
		}
		return c.String(http.StatusInternalServerError, "Error reading page content")
	}

	nonce, _ := c.Get("cspNonce").(string)

	tmpl, err := template.ParseFiles("web/templates/layout.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error loading layout")
	}

	data := struct {
		Content template.HTML
		ID      string
		Nonce   string
	}{
		Content: template.HTML(contentBytes),
		ID:      id,
		Nonce:   nonce,
	}

	return tmpl.Execute(c.Response().Writer, data)
}
