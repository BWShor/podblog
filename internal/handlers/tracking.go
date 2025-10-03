package handlers

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

func PlayHandler(c echo.Context) error {
	id := c.Param("id")
	platform := c.QueryParam("platform")

	log.Printf("Audio %s played on platform=%s\n", id, platform)

	audioPath := filepath.Join("web/content", id, id+".mp3")
	if _, err := filepath.Abs(audioPath); err != nil {
		return c.String(http.StatusNotFound, "Audio not found")
	}

	return c.File(audioPath)
}
