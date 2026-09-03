# Model Context Protocol (MCP) Guide for `agentdoc`

`agentdoc` includes a native, pure-Go **Model Context Protocol (MCP)** JSON-RPC 2.0 stdio server. It allows any LLM agent or MCP client to automatically discover and invoke all document, code, spreadsheet, and image manipulation tools.

---

## 🚀 Quick Setup

### Standard MCP Configuration (`mcp_config.json`)

```json
{
  "mcpServers": {
    "agentdoc": {
      "command": "agentdoc",
      "args": ["mcp"],
      "description": "High-performance CLI & MCP document/code toolkit for AI agents. Inspects, searches, edits, converts, and visually snapshots documents (PDF, DOCX, XLSX, CSV, Markdown, code, and images)."
    }
  }
}
```

---

## 📂 Client Configuration Paths

| Client | Configuration File Path |
| :--- | :--- |
| **Claude Desktop (Windows)** | `%APPDATA%\Claude\claude_desktop_config.json` |
| **Claude Desktop (macOS)** | `~/Library/Application Support/Claude/claude_desktop_config.json` |
| **Antigravity IDE / Gemini** | `%USERPROFILE%\.gemini\config\mcp_config.json` |
| **Cursor** | Settings $\rightarrow$ Features $\rightarrow$ MCP Servers $\rightarrow$ Add New (`agentdoc mcp`) |
| **VS Code (Cline)** | `%APPDATA%\Code\User\globalStorage\saoudrizwan.claude-dev\settings\cline_mcp_settings.json` |
| **VS Code (Roo Code)** | `%APPDATA%\Code\User\globalStorage\rooveterinaryinc.roo-cline\settings\cline_mcp_settings.json` |

---

## 🛠️ Available MCP Tools

### 1. Code & Text Search
- **`search_code`**: Multi-threaded regex or exact text search across files.
  - Arguments: `query` (string), `dir` (string), `regex` (bool), `caseSensitive` (bool), `ext` (string, e.g. "js,ts,go"), `context` (int).
- **`search_files`**: Fast recursive directory crawler matching glob patterns.
  - Arguments: `pattern` (string), `dir` (string), `depth` (int).

### 2. Surgical Text Editing
- **`read_lines`**: Read exact lines with 1-based indexing and context lines.
  - Arguments: `filePath` (string), `start` (int), `end` (int), `context` (int).
- **`replace_lines`**: Atomic line replacement via temp file swap (avoids file corruption).
  - Arguments: `filePath` (string), `start` (int), `end` (int), `content` (string).
- **`insert_lines`**: Insert text before or after any line number.
  - Arguments: `filePath` (string), `line` (int), `content` (string), `before` (bool).
- **`delete_lines`**: Delete specific line numbers or ranges.
  - Arguments: `filePath` (string), `start` (int), `end` (int).
- **`concat_files`**: Concatenate files with custom headers and delimiters.
  - Arguments: `patterns` (array), `output` (string), `header` (string), `delimiter` (string).
- **`clean_file`**: Strip ANSI escape sequences, remove blank lines, trim trailing whitespace, or regex substitution.
  - Arguments: `filePath` (string), `stripAnsi` (bool), `stripBlank` (bool), `trim` (bool), `pattern` (string), `replace` (string).

### 3. Spreadsheets & Tabular Data
- **`sheets_info`**: Inspect sheet names, row counts, and column headers.
  - Arguments: `filePath` (string).
- **`sheets_get_cell`**: Read a cell's string value (e.g. `B4`).
  - Arguments: `filePath` (string), `cell` (string), `sheet` (string).
- **`sheets_set_cell`**: Update a cell's value without rewriting the entire workbook.
  - Arguments: `filePath` (string), `cell` (string), `value` (string), `sheet` (string).
- **`sheets_add_row`**: Append a new row of values.
  - Arguments: `filePath` (string), `values` (array of strings), `sheet` (string).

### 4. Word DOCX Documents
- **`docs_read`**: Extract structured Markdown or plain text.
  - Arguments: `filePath` (string), `asMarkdown` (bool).
- **`docs_snapshot`**: Render visual PNG preview of `.docx` documents for multimodal vision models.
  - Arguments: `filePath` (string), `outputPng` (string).

### 5. PDF Documents
- **`pdf_read`**: Extract text (full or page-by-page).
  - Arguments: `filePath` (string), `byPage` (bool).
- **`pdf_snapshot`**: Render high-resolution visual PNG screenshots of individual, range, or all PDF pages in parallel.
  - Arguments: `filePath` (string), `outputTarget` (string), `page` (int), `all` (bool), `from` (int), `to` (int).

### 6. Media & Image Extraction
- **`image_extract`**: Extract all embedded images from `.pdf`, `.docx`, or `.xlsx` documents.
  - Arguments: `documentPath` (string), `outputDir` (string), `page` (int).

### 7. Universal Converter
- **`convert`**: Universal cross-format converter between Markdown, DOCX, PDF, XLSX, CSV, and images.
  - Arguments: `input` (string), `output` (string), `sheet` (string).

---

## 📚 Built-in MCP Resources

Any agent supporting MCP Resources can directly read embedded documentation:
- `agentdoc://docs/manual`: Full manual of all commands and options.
- `agentdoc://docs/cheatsheet`: Quick cheatsheet of agent tool usage recipes.

## 💡 Built-in MCP Prompts
- `document_agent_instructions`: Pre-baked system instructions for document analysis and editing.
