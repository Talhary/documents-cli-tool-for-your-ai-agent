# `agentdoc` 🤖📄
### High-Performance Concurrent CLI Suite Built for AI Agents

[![CI](https://github.com/Talhary/documents-cli-tool-for-your-ai-agent/actions/workflows/ci.yml/badge.svg)](https://github.com/Talhary/documents-cli-tool-for-your-ai-agent/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/Talhary/documents-cli-tool-for-your-ai-agent?include_prereleases&style=flat-square)](https://github.com/Talhary/documents-cli-tool-for-your-ai-agent/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Talhary/documents-cli-tool-for-your-ai-agent?style=flat-square)](https://golang.org)
[![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-blue?style=flat-square)](#download--platforms)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](LICENSE)

`agentdoc` is an all-in-one, pure-Go command-line tool suite engineered specifically for **AI agents** (LangChain, AutoGPT, CrewAI, Claude Code, Antigravity, OpenAI Assistants) to inspect, manipulate, search, edit, convert, and snapshot documents, spreadsheets, PDFs, code files, and images.

---

## ⚡ Why `agentdoc` for AI Agents?

1. **Structured Machine-Readable Output (`--json`)**: Every command supports a `--json` flag, returning structured JSON envelopes with execution telemetry (`duration_ms`, `files_processed`, `lines_affected`), eliminating brittle text-scraping hacks in LLM tool calling.
2. **Go Concurrency**: Powered by channel-based worker pools that saturate available CPU cores for multi-threaded directory walking, regex code searching, and batch page snapshotting.
3. **Pure-Go Architecture (Zero Dependencies)**: Compiles into a single standalone binary. No Python, LibreOffice, Node.js, or shared C libraries required on the host system.
4. **Multimodal Visual Snapshots**: Renders formatted PNG preview screenshots of PDF pages (`--page 1` or `--all` pages concurrently) and Word DOCX documents so vision-enabled models (e.g. Gemini 1.5/2.0, GPT-4o) can visually analyze document layouts.
5. **Surgical Precision**: Exact 1-based line reading and atomic line replacements via temporary file swap to eliminate file corruption risks.

---

## 📦 Download & Platforms

Pre-compiled standalone binaries are automatically built and released on every release tag for all major platforms and architectures:

| Platform | Architecture | Binary / Archive |
| :--- | :--- | :--- |
| **Windows** | x86_64 (`amd64`) | `agentdoc-vX.Y.Z-windows-amd64.zip` |
| **Windows** | ARM64 (`arm64`) | `agentdoc-vX.Y.Z-windows-arm64.zip` |
| **Windows** | 32-bit (`386`) | `agentdoc-vX.Y.Z-windows-386.zip` |
| **Linux** | x86_64 (`amd64`) | `agentdoc-vX.Y.Z-linux-amd64.tar.gz` |
| **Linux** | ARM64 (`arm64`) | `agentdoc-vX.Y.Z-linux-arm64.tar.gz` |
| **macOS** | Apple Silicon (`arm64`) | `agentdoc-vX.Y.Z-darwin-arm64.tar.gz` |
| **macOS** | Intel (`amd64`) | `agentdoc-vX.Y.Z-darwin-amd64.tar.gz` |

Download the latest release for your platform from [GitHub Releases](https://github.com/Talhary/documents-cli-tool-for-your-ai-agent/releases).

### Install via Go
```bash
go install github.com/Talhary/documents-cli-tool-for-your-ai-agent@latest
```

### Build from Source
```bash
git clone https://github.com/Talhary/documents-cli-tool-for-your-ai-agent.git
cd documents-cli-tool-for-your-ai-agent
go build -ldflags="-s -w" -o agentdoc .
```

---

## 🔌 Model Context Protocol (MCP) Integration

`agentdoc` includes a built-in native **MCP Server** (`agentdoc mcp`) over standard I/O (stdio) using JSON-RPC 2.0. Any MCP-compatible AI agent or IDE can automatically discover and use all document tools without writing any wrappers.

### Configuration (`mcp_config.json` / `claude_desktop_config.json`)

Add the following to your AI client's MCP configuration:

```json
{
  "mcpServers": {
    "agentdoc": {
      "command": "agentdoc",
      "args": ["mcp"]
    }
  }
}
```

#### Client Configuration Paths:
- **Claude Desktop (Windows)**: `%APPDATA%\Claude\claude_desktop_config.json`
- **Claude Desktop (macOS)**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Antigravity IDE / Gemini IDE**: `%USERPROFILE%\.gemini\config\mcp_config.json`
- **Cursor**: Settings $\rightarrow$ Features $\rightarrow$ MCP Servers $\rightarrow$ Add New MCP Server (`Command: agentdoc`, `Args: mcp`)
- **VS Code (Cline / Roo Code)**: `%APPDATA%\Code\User\globalStorage\saoudrizwan.claude-dev\settings\cline_mcp_settings.json`

Once configured, your AI assistant will have direct access to tools like `search_code`, `read_lines`, `replace_lines`, `docs_read`, `docs_snapshot`, `pdf_read`, `pdf_snapshot`, `sheets_set_cell`, `image_extract`, `convert`, and more!

---

## 📖 Command Reference

### Global Flags
| Flag | Short | Default | Description |
| :--- | :--- | :--- | :--- |
| `--json` | | `false` | Return response as structured JSON |
| `--silent` | | `false` | Suppress non-essential progress output |
| `--workers` | `-w` | `NumCPU` | Number of concurrent worker goroutines |

> [!TIP]
> In Windows PowerShell, run `./agentdoc.exe` or `.\agentdoc.exe` when executing from the current directory.

---

### 1. Code & Folder Search (`agentdoc search`)

Multi-threaded regex and exact string search with automatic binary file skipping and context lines.

```bash
# Search for function pattern in JS/TS/Go files with 2 lines of context
agentdoc search code "handleRequest" --dir ./src --ext js,ts,go -C 2

# Case-sensitive regex search
agentdoc search code "const [A-Z_]+ = \d+;" --regex -s

# Return matches formatted as JSON for AI agents
agentdoc search code "TODO" --dir . --json

# Search inside documents (PDF, DOCX, XLSX, CSV, Markdown, text)
agentdoc search docs "billing invoice" --dir ./files --ext pdf,docx -C 1

# Case-sensitive regex search inside documents
agentdoc search docs "Invoice #[0-9]+" --dir ./files --regex

# Find files matching glob patterns recursively
agentdoc search files "*.test.js" --dir . --depth 3
```

#### JSON Response Schema:
```json
{
  "success": true,
  "command": "search.code",
  "data": {
    "query": "handleRequest",
    "is_regex": false,
    "total_matches": 1,
    "files_scanned": 42,
    "files_matched": 1,
    "duration_ms": 5,
    "matches": [
      {
        "file_path": "src/server.go",
        "line_number": 24,
        "column_number": 10,
        "line_text": "func handleRequest(w http.ResponseWriter, r *http.Request) {",
        "matched_text": "handleRequest",
        "before_context": ["// Handles HTTP requests", "type Server struct{}"],
        "after_context": ["    w.WriteHeader(200)"]
      }
    ]
  },
  "stats": {
    "duration_ms": 5,
    "files_processed": 42,
    "matches_found": 1
  }
}
```

---

### 2. Surgical Text Operations & Token Analysis (`agentdoc text`)

Fast, atomic line-based reading, replacing, inserting, concatenating, cleaning, and RAG chunking.

```bash
# Read exact lines 10 to 25
agentdoc text read-lines ./src/index.js --start 10 --end 25 --json

# Atomically replace line 15 with new content
agentdoc text replace-lines ./src/index.js --start 15 --end 15 --content "const port = 3000;"

# Insert lines before line 20
agentdoc text insert-lines ./src/index.js --line 20 --content "// Authenticate request" --before

# Delete lines 40 to 45
agentdoc text delete-lines ./src/index.js --start 40 --end 45

# Concatenate files with custom headers and delimiters for LLM context injection
agentdoc text concat ./src/*.js --output bundle.txt --header "=== File: %b (%n lines) ===" --delimiter "---"

# Strip ANSI codes, remove trailing whitespace, and clean blank lines
agentdoc text clean output.log --strip-ansi --strip-blank --trim

# Estimate LLM token counts for context window budget
agentdoc text tokens report.pdf

# Split file into token-bounded chunks with overlap for RAG pipelines
agentdoc text chunk document.pdf --max-tokens 512 --overlap 50 --json
```

---

### 3. Spreadsheets & Tabular Data (`agentdoc sheets`)

Inspect, convert, and surgically edit Excel `.xlsx` and `.csv` files.

```bash
# Inspect sheet structure, row counts, and column headers
agentdoc sheets info sales.xlsx --json

# Convert XLSX sheet to CSV
agentdoc sheets xlsx2csv sales.xlsx sales.csv --sheet "Q1" --delimiter ","

# Convert CSV to styled XLSX
agentdoc sheets csv2xlsx data.csv data.xlsx --sheet "Report"

# Read a specific cell value
agentdoc sheets get-cell sales.xlsx --cell "B4"

# Update a cell value
agentdoc sheets set-cell sales.xlsx --cell "B4" --value "999.50"

# Append a row of values to a worksheet
agentdoc sheets add-row sales.xlsx --values "2026-03-01,150.00,Approved"
```

---

### 4. Word DOCX Documents (`agentdoc docs`)

Read, edit, and snapshot Microsoft Word `.docx` documents.

```bash
# Inspect document statistics (paragraphs, word counts)
agentdoc docs info report.docx --json

# Extract structured Markdown
agentdoc docs read report.docx --markdown

# In-place text search-and-replace across paragraphs and tables
agentdoc docs edit replace report.docx --search "Draft" --replace "Final" -o final.docx

# Append a heading and paragraph
agentdoc docs edit append report.docx --text "Conclusion and Next Steps" --heading 2

# Visual snapshot to PNG for multimodal AI vision models
agentdoc docs snapshot report.docx preview.png
```

---

### 5. PDF Documents (`agentdoc pdf`)

Read plain text, split, merge, and snapshot PDF pages.

```bash
# Extract full plain text
agentdoc pdf read manual.pdf

# Extract text page-by-page as JSON
agentdoc pdf read manual.pdf --pages --json

# Merge multiple PDFs into one document
agentdoc pdf merge combined.pdf part1.pdf part2.pdf

# Split a page range into a new PDF
agentdoc pdf split book.pdf chapter1.pdf --from 1 --to 15

# Visual snapshot of an individual page (e.g. page 2)
agentdoc pdf snapshot manual.pdf page2.png --page 2

# Visual snapshot of a page range (e.g. pages 1 to 5)
agentdoc pdf snapshot manual.pdf ./images/pages/ --from 1 --to 5

# Visual snapshot of ALL pages concurrently in parallel
agentdoc pdf snapshot manual.pdf ./images/all_pages/ --all

# Visual snapshot of all pages with JSON response for vision agents
agentdoc pdf snapshot manual.pdf ./images/all_pages/ --all --json
```

---

### 6. Image Operations (`agentdoc image`)

Inspect, convert, compile images to PDF, and extract embedded images from documents.

```bash
# Inspect image dimensions & ratio
agentdoc image info photo.png --json

# Convert image format
agentdoc image convert photo.png photo.jpg --quality 90

# Compile multiple images into a multi-page PDF
agentdoc image to-pdf album.pdf scan1.png scan2.jpg scan3.png

# Extract all embedded images from PDF, DOCX, or XLSX
agentdoc image extract document.pdf ./extracted_images/
agentdoc image extract report.docx ./extracted_images/
```

---

### 7. Universal Document Converter (`agentdoc convert`)

Auto-detects formats by file extension and executes conversions:

```bash
# Markdown to DOCX or PDF
agentdoc convert readme.md readme.docx
agentdoc convert readme.md readme.pdf

# DOCX to Markdown or PDF
agentdoc convert report.docx report.md
agentdoc convert report.docx report.pdf

# PDF to DOCX, Markdown, or Plain Text
agentdoc convert document.pdf document.docx
agentdoc convert document.pdf document.md
agentdoc convert document.pdf document.txt

# XLSX to CSV or CSV to XLSX
agentdoc convert data.xlsx data.csv
agentdoc convert data.csv data.xlsx

# Image to PDF
agentdoc convert chart.png chart.pdf
```

---

### 8. Extract Structured Data (`agentdoc extract`)

Extract links, emails, dates, tables, and document metadata across all supported formats:

```bash
# Extract unique URLs, email addresses, and dates
agentdoc extract links report.pdf --json

# Extract tables from DOCX, XLSX, or CSV as structured rows
agentdoc extract tables spreadsheet.xlsx --json

# Extract file metadata (pages, sheets, word counts, timestamps)
agentdoc extract metadata document.docx --json
```

---

### 9. LLM Tool Schemas (`agentdoc schema`)

Export tool schemas formatted for OpenAI Function Calling, Anthropic Claude Tool Use, or MCP:

```bash
# Export all tools for OpenAI Function Calling
agentdoc schema --format openai

# Export single tool for Anthropic Claude Tool Use
agentdoc schema --format anthropic --tool read_lines

# Export standard MCP tool definitions
agentdoc schema --format mcp
```

---

### 10. Batch Manifest Runner (`agentdoc batch`)

Run sequential multi-step tool pipelines from a JSON manifest file:

```bash
agentdoc batch manifest.json
agentdoc batch manifest.json --json --stop-on-error
```

#### Manifest Example (`manifest.json`):
```json
[
  {
    "id": "find_invoices",
    "tool": "search_docs",
    "arguments": {
      "query": "invoice",
      "dir": "./files"
    }
  },
  {
    "id": "extract_contact_info",
    "tool": "extract_links",
    "arguments": {
      "filePath": "./files/report.pdf"
    }
  }
]
```

---

## 🛠️ GitHub Actions CI/CD

The repository includes two specialized GitHub Actions workflows:

1. **Continuous Integration (`.github/workflows/ci.yml`)**:
   - Tests every commit and pull request on Linux, Windows, and macOS.
   - Runs `go test -v -race ./...` across Go 1.24 and Go 1.25.
   - Verifies dependency integrity with `go mod verify`.

2. **Automated Cross-Platform Releases (`.github/workflows/release.yml`)**:
   - Automatically triggered on tag push (e.g., `git push origin v1.0.0`).
   - Cross-compiles for 7 platforms and architectures:
     - `linux/amd64`, `linux/arm64`
     - `darwin/amd64`, `darwin/arm64`
     - `windows/amd64`, `windows/arm64`, `windows/386`
   - Generates SHA-256 checksums (`checksums.txt`).
   - Creates GitHub Releases with downloadable archives.

---

## 🧪 Testing

```bash
# Run unit and integration tests
go test -v ./...

# Run without cache
go test -count=1 ./...

# Run race detector
go test -race ./pkg/...
```

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).
