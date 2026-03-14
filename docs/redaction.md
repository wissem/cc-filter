# Smart file redaction

Smart file redaction is optional and **disabled by default**.

When enabled, cc-filter scans selected files for secrets. If matches are found, it creates a redacted copy and instructs Claude to read that copy.

## Enable redaction

Add this to user or project config:

```yaml
redact_files:
  extensions:
    - ".swift"
    - ".ts"
    - ".go"
    - ".py"
    - ".json"
    - ".yaml"
  filename_patterns:
    - "config"
    - "settings"
    - "secrets"
```

## Fields

- `extensions`: scan files with these extensions
- `filename_patterns`: scan files whose filename contains any listed substring

If both are empty/missing, redaction is disabled.

## Output location

Redacted files are written under:
- `/tmp/claude/redacted/`

They are cleaned up on `SessionEnd` hook.

## Redacted file header

Redacted copies include a short header noting that values were filtered and showing original file path.
