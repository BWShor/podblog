package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"podblog/internal/csp"
	"podblog/internal/handlers"
	"podblog/internal/rss"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Use human-friendly console output
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(csp.Middleware())

	e.Static("/static", "static")
	// e.File("/", "web/templates/layout.html")

	// Routes
	e.GET("/menu", handlers.MenuHandler)
	e.GET("/", handlers.PageFullHandler)
	e.GET("/page/:id", handlers.PageFullHandler)
	e.GET("/page/:id/content", handlers.PageContentHandler)
	e.GET("/play/:id", handlers.PlayHandler)
	e.GET("/rss", rss.RSSHandler)

	log.Info().Msg("Starting server on :8080")

	// Watch docs folder
	go watchDocs("docs")

	if err := e.Start(":8080"); err != nil {
		log.Fatal().Err(err).Msg("Echo server failed")
	}
}

func watchDocs(dir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create file watcher")
	}
	defer watcher.Close()

	if err := watcher.Add(dir); err != nil {
		log.Fatal().Err(err).Msg("Failed to watch directory")
	}

	log.Info().Str("dir", dir).Msg("Watching docs for changes...")

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Create|fsnotify.Write) != 0 && filepath.Ext(event.Name) == ".docx" {
				log.Info().Str("file", event.Name).Msg("Detected .docx change")
				go convertDocx(event.Name)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Error().Err(err).Msg("Watcher error")
		}
	}
}

func convertDocx(path string) {
	log.Info().Str("file", path).Msg("Converting docx")
	cmd := exec.Command("go", "run", "cmd/docx_to_html/main.go", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Error().Err(err).Str("file", path).Msg("Error converting docx")
		return
	}
	log.Info().Str("file", path).Msg("Conversion complete")
}
