package hooks

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"cc-filter/internal/rules"
)

const redactCacheDir = "/tmp/claude/redacted"

type ClaudeHookProcessor struct {
	rules *rules.Rules
}

func NewClaudeHookProcessor(rules *rules.Rules) *ClaudeHookProcessor {
	return &ClaudeHookProcessor{
		rules: rules,
	}
}

func (c *ClaudeHookProcessor) CanHandle(input map[string]interface{}) bool {
	hookEvent, exists := input["hook_event_name"]
	if !exists {
		return false
	}

	switch hookEvent.(string) {
	case "PreToolUse", "UserPromptSubmit", "SessionEnd":
		return true
	default:
		return false
	}
}

func (c *ClaudeHookProcessor) Process(input map[string]interface{}) (string, error) {
	hookEvent := input["hook_event_name"].(string)

	switch hookEvent {
	case "PreToolUse":
		return c.processPreToolUse(input)
	case "UserPromptSubmit":
		return c.processUserPromptSubmit(input)
	case "SessionEnd":
		return c.processSessionEnd(input)
	default:
		if originalJSON, err := json.Marshal(input); err == nil {
			return string(originalJSON), nil
		}
		return "{}", nil
	}
}

func (c *ClaudeHookProcessor) processPreToolUse(input map[string]interface{}) (string, error) {
	toolName, _ := input["tool_name"].(string)
	toolInput, _ := input["tool_input"].(map[string]interface{})

	switch toolName {
	case "Read":
		return c.handleReadTool(toolInput)
	case "Bash":
		return c.handleBashTool(toolInput)
	case "Grep", "Search":
		return c.handleGrepTool(toolInput)
	case "Glob":
		return c.handleGlobTool(toolInput)
	default:
		return c.allowTool()
	}
}

func (c *ClaudeHookProcessor) handleReadTool(toolInput map[string]interface{}) (string, error) {
	filePath, _ := toolInput["file_path"].(string)

	// Allow reads from redacted cache directory
	if strings.HasPrefix(filePath, redactCacheDir) {
		return c.allowTool()
	}

	// Check if file should be completely blocked (e.g., .env files)
	if shouldBlock, reason := c.rules.ShouldBlockFile(filePath); shouldBlock {
		return c.denyTool(reason)
	}

	// Check if file should be redacted (code files that might contain secrets)
	if c.shouldRedactFile(filePath) {
		redactedPath, wasRedacted, err := c.createRedactedFile(filePath)
		if err == nil && wasRedacted {
			// NOTE: updatedInput does NOT work for Read tool file_path (tested Jan 2026)
			// Falling back to deny+redirect which tells Claude to read the redacted file
			return c.denyWithRedirect(filePath, redactedPath)
		}
	}

	return c.allowTool()
}

func (c *ClaudeHookProcessor) handleBashTool(toolInput map[string]interface{}) (string, error) {
	command, _ := toolInput["command"].(string)
	if shouldBlock, reason := c.rules.ShouldBlockCommand(command); shouldBlock {
		return c.denyTool(reason)
	}
	return c.allowTool()
}

func (c *ClaudeHookProcessor) handleGrepTool(toolInput map[string]interface{}) (string, error) {
	pattern, _ := toolInput["pattern"].(string)
	if shouldBlock, reason := c.rules.ShouldBlockSearch(pattern); shouldBlock {
		return c.denyTool(reason)
	}
	return c.allowTool()
}

func (c *ClaudeHookProcessor) handleGlobTool(toolInput map[string]interface{}) (string, error) {
	pattern, _ := toolInput["pattern"].(string)
	for _, blockedPattern := range c.rules.FileBlocks {
		if containsAnywhere(pattern, blockedPattern) {
			return c.denyTool("Pattern may expose sensitive files: " + pattern)
		}
	}
	return c.allowTool()
}

func (c *ClaudeHookProcessor) processUserPromptSubmit(input map[string]interface{}) (string, error) {
	prompt, _ := input["prompt"].(string)
	result := c.rules.FilterContent(prompt)

	// If content was filtered, block and show improved UX
	if result.Filtered {
		// Build detected patterns list
		var patternsDisplay string
		for _, name := range result.MatchedPatterns {
			patternsDisplay += fmt.Sprintf("  • %s\n", name)
		}

		// Copy redacted content to clipboard
		clipboardStatus := "✓ Copied to clipboard - paste to continue"
		if err := copyToClipboard(result.Content); err != nil {
			clipboardStatus = "⚠ Could not copy to clipboard (pbcopy not available)"
		}

		// Also save to file as backup
		c.createRedactedUserInput(prompt, result.Content)

		// Build formatted error message
		separator := "────────────────────────────────────────"
		errorMsg := fmt.Sprintf(
			"⛔ BLOCKED: Sensitive content detected\n\n"+
				"Detected patterns:\n%s\n"+
				"Your message (redacted):\n%s\n%s\n%s\n\n%s",
			patternsDisplay,
			separator,
			result.Content,
			separator,
			clipboardStatus)

		return "", fmt.Errorf(errorMsg)
	}

	// No sensitive content - pass through unchanged
	return "{}", nil
}

// processSessionEnd handles cleanup when Claude Code session ends
func (c *ClaudeHookProcessor) processSessionEnd(input map[string]interface{}) (string, error) {
	// Remove the entire redacted cache directory
	if err := os.RemoveAll(redactCacheDir); err != nil {
		// Log but don't fail - cleanup is best effort
		log.Printf("SessionEnd cleanup warning: %v", err)
	}

	// SessionEnd has no hookSpecificOutput schema - return empty JSON
	return "{}", nil
}

// shouldRedactFile delegates to the rules configuration
func (c *ClaudeHookProcessor) shouldRedactFile(path string) bool {
	return c.rules.ShouldRedactFile(path)
}

// createRedactedFile reads a file, applies redaction, and writes to cache
func (c *ClaudeHookProcessor) createRedactedFile(originalPath string) (string, bool, error) {
	content, err := os.ReadFile(originalPath)
	if err != nil {
		return "", false, err
	}

	filtered := c.rules.FilterContent(string(content))
	if !filtered.Filtered {
		return "", false, nil
	}

	if err := os.MkdirAll(redactCacheDir, 0755); err != nil {
		return "", false, err
	}

	hash := sha256.Sum256([]byte(originalPath))
	cacheName := fmt.Sprintf("%x_%s", hash[:8], filepath.Base(originalPath))
	cachePath := filepath.Join(redactCacheDir, cacheName)

	header := fmt.Sprintf("# ***FILTERED*** REDACTED VERSION - Some sensitive values have been masked\n# Original: %s\n\n", originalPath)
	if err := os.WriteFile(cachePath, []byte(header+filtered.Content), 0644); err != nil {
		return "", false, err
	}

	return cachePath, true, nil
}

// createRedactedUserInput creates a temp file with redacted user input content
func (c *ClaudeHookProcessor) createRedactedUserInput(content string, filteredContent string) (string, error) {
	// Ensure cache directory exists
	if err := os.MkdirAll(redactCacheDir, 0755); err != nil {
		return "", err
	}

	// Generate unique filename using content hash
	hash := sha256.Sum256([]byte(content))
	cacheName := fmt.Sprintf("user_input_%x.txt", hash[:8])
	cachePath := filepath.Join(redactCacheDir, cacheName)

	// Write redacted content with header
	header := "# REDACTED USER INPUT - Sensitive values have been masked\n\n"
	if err := os.WriteFile(cachePath, []byte(header+filteredContent), 0644); err != nil {
		return "", err
	}

	return cachePath, nil
}

// denyWithRedirect blocks the original read and tells Claude to read the redacted version
// DEPRECATED: Use allowWithRedirect for seamless filtering via updatedInput
func (c *ClaudeHookProcessor) denyWithRedirect(originalPath, redactedPath string) (string, error) {
	response := map[string]interface{}{
		"hookSpecificOutput": map[string]interface{}{
			"hookEventName":      "PreToolUse",
			"permissionDecision": "deny",
			"permissionDecisionReason": fmt.Sprintf(
				"SECRETS DETECTED - File contains sensitive data.\n\n"+
					"Original: %s\n\n"+
					"A redacted version has been created. Please read this file instead:\n\n"+
					"    %s",
				originalPath, redactedPath),
		},
	}
	jsonBytes, _ := json.Marshal(response)
	return string(jsonBytes), nil
}

// allowWithRedirect silently redirects the Read tool to the redacted file
// This uses updatedInput to seamlessly filter content without Claude knowing
func (c *ClaudeHookProcessor) allowWithRedirect(redactedPath string) (string, error) {
	response := map[string]interface{}{
		"hookSpecificOutput": map[string]interface{}{
			"hookEventName":      "PreToolUse",
			"permissionDecision": "allow",
			"updatedInput": map[string]interface{}{
				"file_path": redactedPath,
			},
		},
	}
	jsonBytes, _ := json.Marshal(response)
	return string(jsonBytes), nil
}

// allowTool returns a JSON response that approves the tool use
func (c *ClaudeHookProcessor) allowTool() (string, error) {
	response := map[string]interface{}{
		"hookSpecificOutput": map[string]interface{}{
			"hookEventName":      "PreToolUse",
			"permissionDecision": "allow",
		},
	}
	jsonBytes, _ := json.Marshal(response)
	return string(jsonBytes), nil
}

// denyTool returns a JSON response that blocks the tool use
func (c *ClaudeHookProcessor) denyTool(reason string) (string, error) {
	response := map[string]interface{}{
		"hookSpecificOutput": map[string]interface{}{
			"hookEventName":            "PreToolUse",
			"permissionDecision":       "deny",
			"permissionDecisionReason": reason,
		},
	}
	jsonBytes, _ := json.Marshal(response)
	return string(jsonBytes), nil
}

// copyToClipboard copies text to the system clipboard (macOS)
func copyToClipboard(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

func containsAnywhere(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}