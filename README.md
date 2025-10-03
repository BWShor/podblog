# podblog

*A lightweight, Go‑powered static podcast / blog engine*

---

## 🚀 What is podblog?

podblog is a **minimal, fully‑static** web application written in Go that turns Microsoft Word documents into clean, syntax‑highlighted HTML pages and serves them with a single‑page‑app‑style navigation.  
It is built around a few core libraries:

| Library | Purpose |
|---------|---------|
| **Echo** | Fast HTTP framework that powers the API routes and serves static assets. |
| **HTMX** | Enables dynamic, partial page updates without a full reload – the menu, content and audio are swapped into the page on‑the‑fly. |
| **Prism** | Client‑side syntax highlighter for code blocks. |
| **Pandoc** | Converts `.docx` files to HTML, extracting media and allowing raw HTML blocks via a custom Lua filter. |

> **Why this stack?**  
> *Go* gives you a single binary, *Echo* is lightweight and battle‑tested, *HTMX* keeps the UX snappy, *Prism* gives you beautiful code snippets, and *Pandoc* turns your familiar Word docs into web‑ready content without writing any templates.

---

## 📦 Features

| Feature | Description |
|---------|-------------|
| **Dynamic menu** | Scans `web/content/` for folders containing `index.html` and builds a nested navigation tree on‑the‑fly. |
| **Real‑time updates** | The menu is rebuilt on every `/menu` request, so changes to `menuindex.yml` or the content tree are reflected instantly. |
| **CSP with per‑request nonce** | Each request gets a fresh 128‑bit nonce that is injected into inline scripts and styles, keeping the site safe from XSS attacks. |
| **DOCX → HTML conversion** | A background watcher (`fsnotify`) monitors the `docs/` folder. Whenever a `.docx` file is added or modified, a `pandoc` command is run to produce `index.html` (and media) under `web/content/<slug>/`. |
| **Audio streaming** | Simple `/play/:id` endpoint serves MP3 files stored alongside the content. |
| **RSS placeholder** | `/rss` returns a minimal RSS feed – extendable to a real feed generator. |
| **Dockerised dev environment** | A `docker-compose.yml` pulls a Go image, installs `pandoc` & `curl`, and runs the server. |
| **Extensible** | The menu ordering can be overridden by a `menuindex.yml` (YAML, JSON or CSV) – the file is read on each request so edits are live. |

---

## 📁 Project layout

```
podblog/
├─ cmd/
│ ├─ server/ # Echo server + middleware
│ └─ docx_to_html/ # Helper that runs pandoc
├─ internal/
│ ├─ handlers/ # HTTP handlers (menu, page, play, rss)
│ └─ csp/ # CSP nonce middleware
├─ web/
│ ├─ content/ # Generated HTML pages + media
│ ├─ templates/ # layout.html
│ └─ static/ # CSS, JS, Prism assets
├─ docs/ # Source .docx files
├─ menuindex.yml # Optional ordering file
├─ docker-compose.yml
└─ go.mod
```

---

## ⚙️ Prerequisites

| Item | Minimum |
|------|---------|
| **Go** | 1.24+ |
| **Docker** | Optional – used for the dev container and CI |
| **pandoc** | Required for DOCX → HTML conversion (installed automatically in the Docker image) |
| **curl** | Required for downloading htmx and Prism assets (also installed in Docker) |

---

## 🚀 Quick start

  
# 1. Clone the repository  
git clone https://github.com/BWShor/podblog.git  
cd podblog  

# 2. Run locally (requires Go 1.24+)  
go run cmd/server/main.go  

# 3. Open the site  
open http://localhost:8080  

    Tip: If you prefer Docker, just run docker compose up. The container will install the missing tools, build the Go binary and start the server on port 32001 (mapped to 8080 on your host).

📚 Using the system

    Add content
    Place a .docx file in the docs/ folder.
    The watcher will automatically convert it to web/content/<slug>/index.html (and any embedded media).
    The slug is derived from the file name (e.g. My Guide.docx → my-guide).

    Edit the menu
    The menu is generated from the folder structure.
    If you want a custom order, create menuindex.yml in the project root:

   root:  
     - home  
     - about  
     - guides  
     - contact  

   guides:  
     - getting-started  
     - advanced-topics  

The file is read on every /menu request, so you can tweak it live without restarting the server.

    Serve audio
    Drop an MP3 file named <slug>.mp3 into the same folder as the page (e.g. web/content/my-guide/my-guide.mp3).
    Access it via /play/my-guide.

    RSS
    The /rss endpoint currently returns a static placeholder.
    Replace the XML string in internal/rss/feed.go with a dynamic generator that pulls the newest pages from web/content/.

🛠️ Development notes

    Hot‑reload – The watchDocs goroutine uses fsnotify to monitor the docs/ directory.
    It spawns a goroutine that runs go run cmd/docx_to_html/main.go <docx>.
    This keeps the content in sync without manual intervention.

    CSP – The middleware generates a fresh 128‑bit nonce per request, injects it into the HTML via the layout.html template, and sets the Content‑Security‑Policy header accordingly.

    Testing – Run go test ./... to execute any existing unit tests.
    Add more tests for handlers, especially for menu ordering logic.

📄 License

MIT – see the LICENSE file.
👤 Credits

This Readme was built by the GPT‑OSS 20B model running in FP16 on a pair of RTX 3080 GPUs.
