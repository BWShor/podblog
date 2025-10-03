package rss

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func RSSHandler(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/rss+xml")
	rss := `<?xml version="1.0" encoding="UTF-8" ?>
<rss version="2.0">
<channel>
<title>My Podcast</title>
<link>http://localhost:8080/</link>
<description>Example podcast feed</description>
</channel>
</rss>`
	return c.String(http.StatusOK, rss)
}
