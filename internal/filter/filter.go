package filter

import (
	"encoding/json"
	"strings"

	"cc-filter/internal/hooks"
	"cc-filter/internal/rules"
)

type Filter struct {
	rules        *rules.Rules
	hookRegistry *hooks.Registry
}

func New() (*Filter, error) {
	r, err := rules.LoadRules()
	if err != nil {
		return nil, err
	}

	registry := hooks.NewRegistry()
	registry.Register(hooks.NewClaudeHookProcessor(r))

	return &Filter{
		rules:        r,
		hookRegistry: registry,
	}, nil
}

type ProcessResult struct {
	Output   string
	Filtered bool
	Error    error // For blocking errors (exit code 2)
}

func (f *Filter) Process(input string) ProcessResult {
	input = strings.TrimSpace(input)

	if strings.HasPrefix(input, "{") && strings.HasSuffix(input, "}") {
		var hookData map[string]interface{}
		if err := json.Unmarshal([]byte(input), &hookData); err == nil {
			if result, handled, hookErr := f.hookRegistry.Process(hookData); handled {
				if hookErr != nil {
					return ProcessResult{Output: "", Filtered: true, Error: hookErr}
				}
				return ProcessResult{Output: result, Filtered: true, Error: nil}
			}
		}
	}

	result := f.rules.FilterContent(input)
	return ProcessResult{Output: result.Content, Filtered: result.Filtered, Error: nil}
}