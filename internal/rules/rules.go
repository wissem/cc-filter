package rules

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Rules struct {
	Patterns      []PatternRule `yaml:"patterns"`
	FileBlocks    []string      `yaml:"file_blocks"`
	SearchBlocks  []string      `yaml:"search_blocks"`
	CommandBlocks []string      `yaml:"command_blocks"`
	RedactFiles   RedactFiles   `yaml:"redact_files"`

	// compiled regex patterns
	compiledPatterns      []*regexp.Regexp
	compiledCommandBlocks []*regexp.Regexp
}

type PatternRule struct {
	Name        string `yaml:"name"`
	Regex       string `yaml:"regex"`
	Replacement string `yaml:"replacement"`
}

type RedactFiles struct {
	Extensions       []string `yaml:"extensions"`
	FilenamePatterns []string `yaml:"filename_patterns"`
}

func LoadRules() (*Rules, error) {
	// start with defaults
	defaultRules, err := loadDefaultRules()
	if err != nil {
		return nil, err
	}

	// merge user config if exists
	userConfigPath := getUserConfigPath()
	if data, err := os.ReadFile(userConfigPath); err == nil {
		var userRules Rules
		if err := yaml.Unmarshal(data, &userRules); err == nil {
			defaultRules = mergeRules(defaultRules, &userRules)
		}
	}

	// merge project config if exists
	projectConfigPath := "config.yaml"
	if data, err := os.ReadFile(projectConfigPath); err == nil {
		var projectRules Rules
		if err := yaml.Unmarshal(data, &projectRules); err == nil {
			defaultRules = mergeRules(defaultRules, &projectRules)
		}
	}

	return defaultRules.compile()
}

func loadDefaultRules() (*Rules, error) {
	configPath := "configs/default-rules.yaml"
	data, err := os.ReadFile(configPath)
	if err != nil {
		return getMinimalDefaultRules(), nil
	}

	var rules Rules
	if err := yaml.Unmarshal(data, &rules); err != nil {
		return nil, err
	}

	return &rules, nil
}

func getMinimalDefaultRules() *Rules {
	return &Rules{
		Patterns: []PatternRule{
			{
				Name:        "api_keys",
				Regex:       `(?i)api[_-]?key[s]?\s*[:=]\s*['"]?([a-zA-Z0-9_\-]{20,})['"]?`,
				Replacement: "***FILTERED***",
			},
			{
				Name:        "openai_keys",
				Regex:       `sk-[a-zA-Z0-9]{48}`,
				Replacement: "mask",
			},
		},
		FileBlocks: []string{
			".env", ".env.local", "*.key", "*.pem", "*secret*",
		},
		SearchBlocks: []string{
			"api", "key", "secret", "password", "token",
		},
		CommandBlocks: []string{
			"cat.*env", "printenv", "grep.*secret",
		},
	}
}

func getUserConfigPath() string {
	if homeDir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(homeDir, ".cc-filter", "config.yaml")
	}
	return ""
}

func mergeRules(base *Rules, override *Rules) *Rules {
	result := &Rules{
		Patterns:      make([]PatternRule, 0),
		FileBlocks:    make([]string, 0),
		SearchBlocks:  make([]string, 0),
		CommandBlocks: make([]string, 0),
	}

	patternMap := make(map[string]PatternRule)
	for _, pattern := range base.Patterns {
		patternMap[pattern.Name] = pattern
	}

	for _, pattern := range override.Patterns {
		patternMap[pattern.Name] = pattern
	}

	for _, pattern := range patternMap {
		result.Patterns = append(result.Patterns, pattern)
	}
	result.FileBlocks = mergeStringSlices(base.FileBlocks, override.FileBlocks)
	result.SearchBlocks = mergeStringSlices(base.SearchBlocks, override.SearchBlocks)
	result.CommandBlocks = mergeStringSlices(base.CommandBlocks, override.CommandBlocks)
	result.RedactFiles = mergeRedactFiles(base.RedactFiles, override.RedactFiles)

	return result
}

func mergeRedactFiles(base, override RedactFiles) RedactFiles {
	return RedactFiles{
		Extensions:       mergeStringSlices(base.Extensions, override.Extensions),
		FilenamePatterns: mergeStringSlices(base.FilenamePatterns, override.FilenamePatterns),
	}
}

func mergeStringSlices(base, override []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	
	for _, item := range base {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	for _, item := range override {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

func (r *Rules) compile() (*Rules, error) {
	r.compiledPatterns = make([]*regexp.Regexp, len(r.Patterns))
	for i, pattern := range r.Patterns {
		compiled, err := regexp.Compile(pattern.Regex)
		if err != nil {
			return nil, err
		}
		r.compiledPatterns[i] = compiled
	}

	r.compiledCommandBlocks = make([]*regexp.Regexp, len(r.CommandBlocks))
	for i, pattern := range r.CommandBlocks {
		compiled, err := regexp.Compile(strings.ToLower(pattern))
		if err != nil {
			return nil, err
		}
		r.compiledCommandBlocks[i] = compiled
	}

	return r, nil
}

func (r *Rules) ShouldRedactFile(path string) bool {
	// Empty config = disabled
	if len(r.RedactFiles.Extensions) == 0 && len(r.RedactFiles.FilenamePatterns) == 0 {
		return false
	}

	lowerPath := strings.ToLower(path)

	// Check extensions
	for _, ext := range r.RedactFiles.Extensions {
		if ext != "" && strings.HasSuffix(lowerPath, strings.ToLower(ext)) {
			return true
		}
	}

	// Check filename patterns
	baseName := strings.ToLower(filepath.Base(path))
	for _, pattern := range r.RedactFiles.FilenamePatterns {
		if pattern != "" && strings.Contains(baseName, strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

func (r *Rules) ShouldBlockFile(path string) (bool, string) {
	pathLower := strings.ToLower(path)
	
	for _, pattern := range r.FileBlocks {
		patternLower := strings.ToLower(pattern)
		
		if strings.Contains(pattern, "*") {
			if matched, _ := filepath.Match(patternLower, pathLower); matched {
				return true, "Access denied to sensitive file: " + path
			}
		} else {
			if strings.Contains(pathLower, patternLower) {
				return true, "Access denied to sensitive file: " + path
			}
		}
	}
	
	return false, ""
}

func (r *Rules) ShouldBlockSearch(pattern string) (bool, string) {
	patternLower := strings.ToLower(pattern)
	
	for _, blocked := range r.SearchBlocks {
		if strings.Contains(patternLower, strings.ToLower(blocked)) {
			return true, "Search pattern may expose sensitive data: " + pattern
		}
	}
	
	return false, ""
}

func (r *Rules) ShouldBlockCommand(cmd string) (bool, string) {
	cmdLower := strings.ToLower(cmd)
	
	for _, pattern := range r.compiledCommandBlocks {
		if pattern.MatchString(cmdLower) {
			return true, "Command may expose sensitive data: " + cmd
		}
	}
	
	return false, ""
}

type FilterResult struct {
	Content  string
	Filtered bool
}

func (r *Rules) FilterContent(text string) FilterResult {
	filtered := text
	hasChanged := false

	for i, pattern := range r.compiledPatterns {
		rule := r.Patterns[i]
		original := filtered

		switch rule.Replacement {
		case "mask":
			filtered = pattern.ReplaceAllStringFunc(filtered, func(match string) string {
				return strings.Repeat("*", len(match))
			})
		case "env_filter":
			filtered = pattern.ReplaceAllStringFunc(filtered, func(match string) string {
				parts := strings.SplitN(match, "=", 2)
				if len(parts) == 2 {
					return parts[0] + "=***FILTERED***"
				}
				return match
			})
		default:
			filtered = pattern.ReplaceAllString(filtered, rule.Replacement)
		}

		if filtered != original {
			hasChanged = true
		}
	}

	return FilterResult{Content: filtered, Filtered: hasChanged}
}