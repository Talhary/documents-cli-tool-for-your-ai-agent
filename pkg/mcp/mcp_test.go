package mcp

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestMCPServer_HandshakeAndTools(t *testing.T) {
	inputLines := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"search_code","arguments":{"query":"package","dir":".","ext":"go"}}}`,
	}

	inBuf := bytes.NewBufferString(strings.Join(inputLines, "\n") + "\n")
	outBuf := new(bytes.Buffer)

	err := RunServer(inBuf, outBuf)
	if err != nil {
		t.Fatalf("RunServer error: %v", err)
	}

	responses := strings.Split(strings.TrimSpace(outBuf.String()), "\n")
	if len(responses) != 3 {
		t.Fatalf("expected 3 responses, got %d:\n%s", len(responses), outBuf.String())
	}

	// Verify initialize response
	var initResp JSONRPCResponse
	if err := json.Unmarshal([]byte(responses[0]), &initResp); err != nil {
		t.Fatalf("failed unmarshaling init response: %v", err)
	}
	if initResp.Error != nil {
		t.Errorf("init returned error: %+v", initResp.Error)
	}

	// Verify tools/list response
	var listResp JSONRPCResponse
	if err := json.Unmarshal([]byte(responses[1]), &listResp); err != nil {
		t.Fatalf("failed unmarshaling list response: %v", err)
	}
	resMap, ok := listResp.Result.(map[string]any)
	if !ok {
		t.Fatalf("invalid result type: %+v", listResp.Result)
	}
	tools, ok := resMap["tools"].([]any)
	if !ok || len(tools) == 0 {
		t.Fatalf("expected tools in list, got: %+v", resMap)
	}
}

func TestDispatch_NewTools(t *testing.T) {
	// Test text_tokens
	rawTokens, isErr := Dispatch("text_tokens", []byte(`{"filePath":"../../main.go"}`))
	if isErr {
		// Try local main.go
		rawTokens, isErr = Dispatch("text_tokens", []byte(`{"filePath":"main.go"}`))
	}
	if isErr {
		t.Fatalf("Dispatch text_tokens failed: %s", rawTokens)
	}

	// Test extract_links
	rawLinks, isErr := Dispatch("extract_links", []byte(`{"filePath":"../../README.md"}`))
	if isErr {
		rawLinks, isErr = Dispatch("extract_links", []byte(`{"filePath":"README.md"}`))
	}
	if isErr {
		t.Fatalf("Dispatch extract_links failed: %s", rawLinks)
	}
}

