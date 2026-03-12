package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cc-filter/internal/rules"
)

func testDefaultRules() []byte {
	data, err := os.ReadFile("../../configs/default-rules.yaml")
	if err != nil {
		panic("failed to read default rules for test: " + err.Error())
	}
	return data
}

func TestAllowWithRedirect(t *testing.T) {
	r, _ := rules.LoadRules(testDefaultRules())
	processor := NewClaudeHookProcessor(r)

	redactedPath := "/tmp/claude/redacted/abc123_config.swift"
	result, err := processor.allowWithRedirect(redactedPath)

	if err != nil {
		t.Fatalf("allowWithRedirect returned error: %v", err)
	}

	// Parse the JSON response
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// Verify hookSpecificOutput exists
	hookOutput, ok := response["hookSpecificOutput"].(map[string]interface{})
	if !ok {
		t.Fatal("hookSpecificOutput not found or wrong type")
	}

	// Verify hookEventName
	if hookOutput["hookEventName"] != "PreToolUse" {
		t.Errorf("hookEventName = %v, want PreToolUse", hookOutput["hookEventName"])
	}

	// Verify permissionDecision is "allow"
	if hookOutput["permissionDecision"] != "allow" {
		t.Errorf("permissionDecision = %v, want allow", hookOutput["permissionDecision"])
	}

	// Verify updatedInput exists and has file_path
	updatedInput, ok := hookOutput["updatedInput"].(map[string]interface{})
	if !ok {
		t.Fatal("updatedInput not found or wrong type")
	}

	if updatedInput["file_path"] != redactedPath {
		t.Errorf("file_path = %v, want %v", updatedInput["file_path"], redactedPath)
	}
}

func TestAllowTool(t *testing.T) {
	r, _ := rules.LoadRules(testDefaultRules())
	processor := NewClaudeHookProcessor(r)

	result, err := processor.allowTool()

	if err != nil {
		t.Fatalf("allowTool returned error: %v", err)
	}

	if result != "" {
		t.Errorf("allowTool should return empty string, got %q", result)
	}
}

func TestDenyTool(t *testing.T) {
	r, _ := rules.LoadRules(testDefaultRules())
	processor := NewClaudeHookProcessor(r)

	reason := "Access denied to sensitive file"
	result, err := processor.denyTool(reason)

	if err != nil {
		t.Fatalf("denyTool returned error: %v", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	hookOutput := response["hookSpecificOutput"].(map[string]interface{})

	if hookOutput["permissionDecision"] != "deny" {
		t.Errorf("permissionDecision = %v, want deny", hookOutput["permissionDecision"])
	}

	if hookOutput["permissionDecisionReason"] != reason {
		t.Errorf("permissionDecisionReason = %v, want %v", hookOutput["permissionDecisionReason"], reason)
	}
}

func TestHandleReadToolWithSecrets(t *testing.T) {
	// Create a temp file with secrets
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "config.swift")
	content := `let apiKey = "sk-1234567890abcdefghijklmnopqrstuvwxyz123456789012"`
	os.WriteFile(testFile, []byte(content), 0644)

	r, _ := rules.LoadRules(testDefaultRules())
	r.RedactFiles.Extensions = append(r.RedactFiles.Extensions, ".swift")
	processor := NewClaudeHookProcessor(r)

	toolInput := map[string]interface{}{
		"file_path": testFile,
	}

	result, err := processor.handleReadTool(toolInput)

	if err != nil {
		t.Fatalf("handleReadTool returned error: %v", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	hookOutput := response["hookSpecificOutput"].(map[string]interface{})
	decision := hookOutput["permissionDecision"].(string)

	if decision != "deny" {
		t.Errorf("Expected deny for file with secrets, got %v", decision)
	}

	reason := hookOutput["permissionDecisionReason"].(string)
	if !strings.Contains(reason, "/tmp/claude/redacted/") {
		t.Errorf("Deny reason should contain redirect path, got %v", reason)
	}
}

func TestHandleReadToolWithoutSecrets(t *testing.T) {
	// Create a temp file without secrets
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "clean.swift")
	content := `let greeting = "Hello, World!"`
	os.WriteFile(testFile, []byte(content), 0644)

	r, _ := rules.LoadRules(testDefaultRules())
	processor := NewClaudeHookProcessor(r)

	toolInput := map[string]interface{}{
		"file_path": testFile,
	}

	result, err := processor.handleReadTool(toolInput)

	if err != nil {
		t.Fatalf("handleReadTool returned error: %v", err)
	}

	if result != "" {
		t.Errorf("handleReadTool for clean file should return empty string, got %q", result)
	}
}

func TestHandleReadToolBlockedFile(t *testing.T) {
	// Test that .env files are still blocked (not redirected)
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, ".env")
	content := `API_KEY=secret123`
	os.WriteFile(testFile, []byte(content), 0644)

	r, _ := rules.LoadRules(testDefaultRules())
	processor := NewClaudeHookProcessor(r)

	toolInput := map[string]interface{}{
		"file_path": testFile,
	}

	result, err := processor.handleReadTool(toolInput)

	if err != nil {
		t.Fatalf("handleReadTool returned error: %v", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	hookOutput := response["hookSpecificOutput"].(map[string]interface{})

	// .env files should be denied (blocked), not redirected
	if hookOutput["permissionDecision"] != "deny" {
		t.Errorf("permissionDecision = %v, want deny for .env file", hookOutput["permissionDecision"])
	}

	// Should NOT have updatedInput (it's a hard block)
	if _, exists := hookOutput["updatedInput"]; exists {
		t.Error("Blocked .env file should not have updatedInput")
	}
}
