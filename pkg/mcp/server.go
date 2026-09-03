package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"docs-cli/pkg/converter"
	"docs-cli/pkg/docs"
	"docs-cli/pkg/imgops"
	"docs-cli/pkg/pdf"
	"docs-cli/pkg/search"
	"docs-cli/pkg/sheets"
	"docs-cli/pkg/textops"
)

// JSONRPCRequest represents an incoming MCP request.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents an outgoing MCP response.
type JSONRPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Result  any    `json:"result,omitempty"`
	Error   *RPCError `json:"error,omitempty"`
}

// RPCError defines JSON-RPC error.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Tool defines an MCP tool definition.
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

// InputSchema defines parameters schema.
type InputSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]Property    `json:"properties"`
	Required   []string               `json:"required,omitempty"`
}

// Property details a parameter.
type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// ToolCallResult defines content returned to the LLM.
type ToolCallResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ToolContent item.
type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// RunServer runs the stdio MCP server loop.
func RunServer(r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)

	tools := getToolDefinitions()

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			continue
		}

		switch req.Method {
		case "initialize":
			resp := JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]any{
					"protocolVersion": "2024-11-05",
					"capabilities": map[string]any{
						"tools": map[string]any{},
					},
					"serverInfo": map[string]any{
						"name":    "agentdoc",
						"version": "1.0.0",
					},
				},
			}
			sendResponse(w, resp)

		case "notifications/initialized":
			// No-op for notification

		case "ping":
			sendResponse(w, JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  map[string]any{},
			})

		case "tools/list":
			resp := JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]any{
					"tools": tools,
				},
			}
			sendResponse(w, resp)

		case "tools/call":
			var callParams struct {
				Name      string          `json:"name"`
				Arguments json.RawMessage `json:"arguments"`
			}
			if err := json.Unmarshal(req.Params, &callParams); err != nil {
				sendError(w, req.ID, -32602, "Invalid params")
				continue
			}

			resultText, isErr := handleToolCall(callParams.Name, callParams.Arguments)
			resp := JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: ToolCallResult{
					Content: []ToolContent{
						{
							Type: "text",
							Text: resultText,
						},
					},
					IsError: isErr,
				},
			}
			sendResponse(w, resp)

		default:
			if req.ID != nil {
				sendError(w, req.ID, -32601, fmt.Sprintf("Method not found: %s", req.Method))
			}
		}
	}

	return scanner.Err()
}

func sendResponse(w io.Writer, resp JSONRPCResponse) {
	data, _ := json.Marshal(resp)
	fmt.Fprintf(w, "%s\n", string(data))
}

func sendError(w io.Writer, id any, code int, message string) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
	}
	sendResponse(w, resp)
}

func getToolDefinitions() []Tool {
	return []Tool{
		{
			Name:        "search_code",
			Description: "Multi-threaded search for regex or exact text in code/text files with context lines",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"query":         {Type: "string", Description: "Search query or regex"},
					"dir":           {Type: "string", Description: "Root directory (defaults to .)"},
					"regex":         {Type: "boolean", Description: "Treat query as regex"},
					"caseSensitive": {Type: "boolean", Description: "Case-sensitive match"},
					"ext":           {Type: "string", Description: "Comma-separated extensions (e.g. js,ts,go,py)"},
					"context":       {Type: "integer", Description: "Lines of context before/after"},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "search_files",
			Description: "Find files and directories by glob pattern recursively",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"pattern": {Type: "string", Description: "Glob pattern (e.g. *.test.js)"},
					"dir":     {Type: "string", Description: "Root directory"},
					"depth":   {Type: "integer", Description: "Max recursion depth"},
				},
			},
		},
		{
			Name:        "read_lines",
			Description: "Read exact line or range of lines from a readable file with 1-based indexing",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"filePath": {Type: "string", Description: "Path to file"},
					"start":    {Type: "integer", Description: "Starting line (1-based)"},
					"end":      {Type: "integer", Description: "Ending line"},
					"context":  {Type: "integer", Description: "Context lines"},
				},
				Required: []string{"filePath"},
			},
		},
		{
			Name:        "replace_lines",
			Description: "Atomically replace exact line or line range with new content",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"filePath": {Type: "string", Description: "Path to file"},
					"start":    {Type: "integer", Description: "Starting line"},
					"end":      {Type: "integer", Description: "Ending line"},
					"content":  {Type: "string", Description: "Replacement content string"},
				},
				Required: []string{"filePath", "start", "content"},
			},
		},
		{
			Name:        "insert_lines",
			Description: "Insert content before or after a specific line number",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"filePath": {Type: "string", Description: "Path to file"},
					"line":     {Type: "integer", Description: "Target line number"},
					"content":  {Type: "string", Description: "Content to insert"},
					"before":   {Type: "boolean", Description: "Insert before line (default after)"},
				},
				Required: []string{"filePath", "line", "content"},
			},
		},
		{
			Name:        "delete_lines",
			Description: "Delete line or line range from a readable file",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"filePath": {Type: "string", Description: "Path to file"},
					"start":    {Type: "integer", Description: "Starting line"},
					"end":      {Type: "integer", Description: "Ending line"},
				},
				Required: []string{"filePath", "start"},
			},
		},
		{
			Name:        "concat_files",
			Description: "Concatenate multiple text or code files with custom headers and delimiters",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"patterns":  {Type: "array", Description: "List of file paths or globs"},
					"output":    {Type: "string", Description: "Output file path (optional)"},
					"header":    {Type: "string", Description: "Header template (e.g. === %b ===)"},
					"delimiter": {Type: "string", Description: "Delimiter between files"},
				},
				Required: []string{"patterns"},
			},
		},
		{
			Name:        "clean_file",
			Description: "Clean file by stripping ANSI codes, trailing whitespace, blank lines, or regex pattern",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"filePath":   {Type: "string", Description: "Path to file"},
					"stripAnsi":  {Type: "boolean", Description: "Strip ANSI escape sequences"},
					"stripBlank": {Type: "boolean", Description: "Remove blank lines"},
					"trim":       {Type: "boolean", Description: "Trim trailing whitespace"},
					"pattern":    {Type: "string", Description: "Regex pattern to replace"},
					"replace":    {Type: "string", Description: "Replacement text for pattern"},
				},
				Required: []string{"filePath"},
			},
		},
		{
			Name:        "sheets_info",
			Description: "Inspect sheet names, row counts, and column headers of an XLSX spreadsheet",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"filePath": {Type: "string", Description: "Path to .xlsx file"},
				},
				Required: []string{"filePath"},
			},
		},
		{
			Name:        "sheets_set_cell",
			Description: "Update a cell value in an XLSX workbook",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"filePath": {Type: "string", Description: "Path to .xlsx file"},
					"cell":     {Type: "string", Description: "Cell coordinate (e.g. A1, B4)"},
					"value":    {Type: "string", Description: "Value to write"},
					"sheet":    {Type: "string", Description: "Sheet name (optional)"},
				},
				Required: []string{"filePath", "cell", "value"},
			},
		},
		{
			Name:        "sheets_get_cell",
			Description: "Read a cell value from an XLSX workbook",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"filePath": {Type: "string", Description: "Path to .xlsx file"},
					"cell":     {Type: "string", Description: "Cell coordinate (e.g. A1, B4)"},
					"sheet":    {Type: "string", Description: "Sheet name (optional)"},
				},
				Required: []string{"filePath", "cell"},
			},
		},
		{
			Name:        "sheets_add_row",
			Description: "Append a row to an XLSX worksheet",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"filePath": {Type: "string", Description: "Path to .xlsx file"},
					"values":   {Type: "array", Description: "List of string values for row"},
					"sheet":    {Type: "string", Description: "Sheet name (optional)"},
				},
				Required: []string{"filePath", "values"},
			},
		},
		{
			Name:        "docs_read",
			Description: "Read text and structure from a Word DOCX document as Markdown or plain text",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"filePath":   {Type: "string", Description: "Path to .docx file"},
					"asMarkdown": {Type: "boolean", Description: "Format as Markdown (default true)"},
				},
				Required: []string{"filePath"},
			},
		},
		{
			Name:        "docs_snapshot",
			Description: "Render visual preview screenshot of a Word DOCX document to PNG for AI vision models",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"filePath":  {Type: "string", Description: "Path to .docx file"},
					"outputPng": {Type: "string", Description: "Output PNG file path"},
				},
				Required: []string{"filePath", "outputPng"},
			},
		},
		{
			Name:        "pdf_read",
			Description: "Read plain text from a PDF file",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"filePath": {Type: "string", Description: "Path to .pdf file"},
					"byPage":   {Type: "boolean", Description: "Separate text by page"},
				},
				Required: []string{"filePath"},
			},
		},
		{
			Name:        "pdf_snapshot",
			Description: "Render visual preview screenshot of individual, range, or all PDF pages to PNG for vision models",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"filePath":     {Type: "string", Description: "Path to .pdf file"},
					"outputTarget": {Type: "string", Description: "Destination PNG path or directory"},
					"page":         {Type: "integer", Description: "Single page number (default 1)"},
					"all":          {Type: "boolean", Description: "Render all pages in parallel"},
					"from":         {Type: "integer", Description: "Start page range"},
					"to":           {Type: "integer", Description: "End page range"},
				},
				Required: []string{"filePath", "outputTarget"},
			},
		},
		{
			Name:        "image_extract",
			Description: "Extract all embedded images and media from a PDF, DOCX, or XLSX document",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"documentPath": {Type: "string", Description: "Path to .pdf, .docx, or .xlsx file"},
					"outputDir":    {Type: "string", Description: "Directory to save extracted images"},
					"page":         {Type: "integer", Description: "Optional page number for PDF"},
				},
				Required: []string{"documentPath", "outputDir"},
			},
		},
		{
			Name:        "convert",
			Description: "Universal converter between Markdown, Word DOCX, PDF, XLSX, CSV, and images",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"input":  {Type: "string", Description: "Input file path"},
					"output": {Type: "string", Description: "Output file path with target extension"},
					"sheet":  {Type: "string", Description: "Optional sheet name for XLSX"},
				},
				Required: []string{"input", "output"},
			},
		},
	}
}

func handleToolCall(name string, rawArgs json.RawMessage) (string, bool) {
	switch name {
	case "search_code":
		var args struct {
			Query         string `json:"query"`
			Dir           string `json:"dir"`
			Regex         bool   `json:"regex"`
			CaseSensitive bool   `json:"caseSensitive"`
			Ext           string `json:"ext"`
			Context       int    `json:"context"`
		}
		json.Unmarshal(rawArgs, &args)
		var includes []string
		if args.Ext != "" {
			for _, ext := range strings.Split(args.Ext, ",") {
				ext = strings.TrimSpace(ext)
				if !strings.HasPrefix(ext, "*") {
					ext = "*." + strings.TrimPrefix(ext, ".")
				}
				includes = append(includes, ext)
			}
		}
		opts := search.SearchOptions{
			Query:         args.Query,
			RootDir:       args.Dir,
			IsRegex:       args.Regex,
			CaseSensitive: args.CaseSensitive,
			IncludeGlobs:  includes,
			ContextBefore: args.Context,
			ContextAfter:  args.Context,
		}
		res, err := search.SearchCode(context.Background(), opts)
		if err != nil {
			return err.Error(), true
		}
		b, _ := json.MarshalIndent(res, "", "  ")
		return string(b), false

	case "search_files":
		var args struct {
			Pattern string `json:"pattern"`
			Dir     string `json:"dir"`
			Depth   int    `json:"depth"`
		}
		json.Unmarshal(rawArgs, &args)
		var inc []string
		if args.Pattern != "" {
			inc = append(inc, args.Pattern)
		}
		entries, err := search.FindFiles(search.FindOptions{
			RootPath:     args.Dir,
			IncludeGlobs: inc,
			MaxDepth:     args.Depth,
			IgnoreHidden: true,
		})
		if err != nil {
			return err.Error(), true
		}
		b, _ := json.MarshalIndent(entries, "", "  ")
		return string(b), false

	case "read_lines":
		var args struct {
			FilePath string `json:"filePath"`
			Start    int    `json:"start"`
			End      int    `json:"end"`
			Context  int    `json:"context"`
		}
		json.Unmarshal(rawArgs, &args)
		res, err := textops.ReadLines(args.FilePath, args.Start, args.End, args.Context)
		if err != nil {
			return err.Error(), true
		}
		b, _ := json.MarshalIndent(res, "", "  ")
		return string(b), false

	case "replace_lines":
		var args struct {
			FilePath string `json:"filePath"`
			Start    int    `json:"start"`
			End      int    `json:"end"`
			Content  string `json:"content"`
		}
		json.Unmarshal(rawArgs, &args)
		if args.End == 0 {
			args.End = args.Start
		}
		res, err := textops.ReplaceLines(args.FilePath, args.Start, args.End, args.Content)
		if err != nil {
			return err.Error(), true
		}
		b, _ := json.MarshalIndent(res, "", "  ")
		return string(b), false

	case "insert_lines":
		var args struct {
			FilePath string `json:"filePath"`
			Line     int    `json:"line"`
			Content  string `json:"content"`
			Before   bool   `json:"before"`
		}
		json.Unmarshal(rawArgs, &args)
		res, err := textops.InsertLines(args.FilePath, args.Line, args.Content, args.Before)
		if err != nil {
			return err.Error(), true
		}
		b, _ := json.MarshalIndent(res, "", "  ")
		return string(b), false

	case "delete_lines":
		var args struct {
			FilePath string `json:"filePath"`
			Start    int    `json:"start"`
			End      int    `json:"end"`
		}
		json.Unmarshal(rawArgs, &args)
		if args.End == 0 {
			args.End = args.Start
		}
		res, err := textops.DeleteLines(args.FilePath, args.Start, args.End)
		if err != nil {
			return err.Error(), true
		}
		b, _ := json.MarshalIndent(res, "", "  ")
		return string(b), false

	case "concat_files":
		var args struct {
			Patterns  []string `json:"patterns"`
			Output    string   `json:"output"`
			Header    string   `json:"header"`
			Delimiter string   `json:"delimiter"`
		}
		json.Unmarshal(rawArgs, &args)
		res, plain, err := textops.ConcatFiles(args.Patterns, textops.ConcatOptions{
			OutputFile:     args.Output,
			HeaderTemplate: args.Header,
			Delimiter:      args.Delimiter,
		})
		if err != nil {
			return err.Error(), true
		}
		if args.Output != "" {
			b, _ := json.MarshalIndent(res, "", "  ")
			return string(b), false
		}
		return plain, false

	case "clean_file":
		var args struct {
			FilePath   string `json:"filePath"`
			StripAnsi  bool   `json:"stripAnsi"`
			StripBlank bool   `json:"stripBlank"`
			Trim       bool   `json:"trim"`
			Pattern    string `json:"pattern"`
			Replace    string `json:"replace"`
		}
		json.Unmarshal(rawArgs, &args)
		res, err := textops.CleanFile(args.FilePath, textops.CleanOptions{
			StripANSI:    args.StripAnsi,
			StripBlank:   args.StripBlank,
			TrimTrailing: args.Trim,
			RegexPattern: args.Pattern,
			RegexReplace: args.Replace,
		})
		if err != nil {
			return err.Error(), true
		}
		b, _ := json.MarshalIndent(res, "", "  ")
		return string(b), false

	case "sheets_info":
		var args struct {
			FilePath string `json:"filePath"`
		}
		json.Unmarshal(rawArgs, &args)
		info, err := sheets.InspectSheet(args.FilePath)
		if err != nil {
			return err.Error(), true
		}
		b, _ := json.MarshalIndent(info, "", "  ")
		return string(b), false

	case "sheets_set_cell":
		var args struct {
			FilePath string `json:"filePath"`
			Cell     string `json:"cell"`
			Value    string `json:"value"`
			Sheet    string `json:"sheet"`
		}
		json.Unmarshal(rawArgs, &args)
		err := sheets.SetCellValue(args.FilePath, args.Sheet, args.Cell, args.Value)
		if err != nil {
			return err.Error(), true
		}
		return fmt.Sprintf("Successfully set %s cell %s = %s", args.FilePath, args.Cell, args.Value), false

	case "sheets_get_cell":
		var args struct {
			FilePath string `json:"filePath"`
			Cell     string `json:"cell"`
			Sheet    string `json:"sheet"`
		}
		json.Unmarshal(rawArgs, &args)
		val, err := sheets.GetCellValue(args.FilePath, args.Sheet, args.Cell)
		if err != nil {
			return err.Error(), true
		}
		return val, false

	case "sheets_add_row":
		var args struct {
			FilePath string   `json:"filePath"`
			Values   []string `json:"values"`
			Sheet    string   `json:"sheet"`
		}
		json.Unmarshal(rawArgs, &args)
		rowIdx, err := sheets.AddRow(args.FilePath, args.Sheet, args.Values)
		if err != nil {
			return err.Error(), true
		}
		return fmt.Sprintf("Successfully appended row %d to %s", rowIdx, args.FilePath), false

	case "docs_read":
		var args struct {
			FilePath   string `json:"filePath"`
			AsMarkdown bool   `json:"asMarkdown"`
		}
		json.Unmarshal(rawArgs, &args)
		if args.AsMarkdown {
			md, err := docs.DOCXToMarkdown(args.FilePath)
			if err != nil {
				return err.Error(), true
			}
			return md, false
		}
		txt, err := docs.DOCXToText(args.FilePath)
		if err != nil {
			return err.Error(), true
		}
		return txt, false

	case "docs_snapshot":
		var args struct {
			FilePath  string `json:"filePath"`
			OutputPng string `json:"outputPng"`
		}
		json.Unmarshal(rawArgs, &args)
		err := docs.SnapshotDOCX(args.FilePath, args.OutputPng)
		if err != nil {
			return err.Error(), true
		}
		return fmt.Sprintf("Snapshot created at %s", args.OutputPng), false

	case "pdf_read":
		var args struct {
			FilePath string `json:"filePath"`
			ByPage   bool   `json:"byPage"`
		}
		json.Unmarshal(rawArgs, &args)
		if args.ByPage {
			pages, err := pdf.ExtractPages(args.FilePath)
			if err != nil {
				return err.Error(), true
			}
			b, _ := json.MarshalIndent(pages, "", "  ")
			return string(b), false
		}
		txt, err := pdf.ExtractText(args.FilePath)
		if err != nil {
			return err.Error(), true
		}
		return txt, false

	case "pdf_snapshot":
		var args struct {
			FilePath     string `json:"filePath"`
			OutputTarget string `json:"outputTarget"`
			Page         int    `json:"page"`
			All          bool   `json:"all"`
			From         int    `json:"from"`
			To           int    `json:"to"`
		}
		json.Unmarshal(rawArgs, &args)
		if args.All || args.From > 0 || args.To > 0 {
			files, err := pdf.SnapshotAllPages(args.FilePath, args.OutputTarget, args.From, args.To, 0)
			if err != nil {
				return err.Error(), true
			}
			b, _ := json.MarshalIndent(files, "", "  ")
			return string(b), false
		}
		err := pdf.SnapshotPDF(args.FilePath, args.OutputTarget, args.Page)
		if err != nil {
			return err.Error(), true
		}
		return fmt.Sprintf("Snapshot of page %d created at %s", args.Page, args.OutputTarget), false

	case "image_extract":
		var args struct {
			DocumentPath string `json:"documentPath"`
			OutputDir    string `json:"outputDir"`
			Page         int    `json:"page"`
		}
		json.Unmarshal(rawArgs, &args)
		files, err := imgops.ExtractMedia(args.DocumentPath, args.OutputDir, args.Page)
		if err != nil {
			return err.Error(), true
		}
		b, _ := json.MarshalIndent(files, "", "  ")
		return string(b), false

	case "convert":
		var args struct {
			Input     string `json:"input"`
			Output    string `json:"output"`
			Sheet     string `json:"sheet"`
			Delimiter string `json:"delimiter"`
		}
		json.Unmarshal(rawArgs, &args)
		var delim rune = ','
		if args.Delimiter != "" {
			delim = rune(args.Delimiter[0])
		}
		err := converter.Convert(args.Input, args.Output, converter.ConvertOptions{
			SheetName: args.Sheet,
			Delimiter: delim,
		})
		if err != nil {
			return err.Error(), true
		}
		return fmt.Sprintf("Successfully converted %s -> %s", args.Input, args.Output), false

	default:
		return fmt.Sprintf("Tool not found: %s", name), true
	}
}
