package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Use human-friendly console writer for stdout
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	if len(os.Args) < 2 {
		log.Error().Msg("Usage: docx_to_html <path/to/docx>")
		os.Exit(1)
	}

	docxPath := os.Args[1]
	baseName := filepath.Base(docxPath)
	name := baseName[:len(baseName)-len(filepath.Ext(baseName))]

	// Normalize name: replace spaces with "-", convert to lowercase
	normalized := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	outDir := filepath.Join("web/content", normalized)

	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatal().Err(err).Msg("Failed to create output directory")
	}

	log.Info().
		Str("docx", docxPath).
		Str("outputDir", outDir).
		Msg("Starting conversion")
	luaFilter := filepath.Join("cmd", "docx_to_html", "inline_html.lua")

	cmd := exec.Command("pandoc", docxPath,
		"-o", filepath.Join(outDir, "index.html"),
		"--extract-media", filepath.Join(outDir, "media"),
		"--lua-filter", luaFilter,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal().Err(err).Msg("Pandoc conversion failed")
	}

	log.Info().Msg("Conversion complete")
}
