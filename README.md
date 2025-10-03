# podblog

A minimal, Go‑based static podcast/blog site that uses **Echo**, **HTMX**, **Prism**, and **Pandoc** to convert DOCX to HTML.

## Features

- Dynamic menu generation from `web/content/`
- HTMX-powered navigation (no full page reloads)
- CSP with per‑request nonces
- Automatic DOCX → HTML conversion via `pandoc`
- RSS feed (placeholder)

## Prerequisites

- Go 1.24+
- Docker (optional, for dev container)
- `pandoc` and `curl` (for the Docker image)

## Quick start

```bash  
# Clone  
git clone https://github.com/You/podblog.git  
cd podblog  

# Run locally  
go run cmd/server/main.go  
open http://localhost:8080
