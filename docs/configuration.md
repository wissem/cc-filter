# Configuration reference

cc-filter loads configuration from three layers (lowest to highest priority):

1. `configs/default-rules.yaml` (built-in)
2. `~/.cc-filter/config.yaml` (user)
3. `./config.yaml` (project)

Later layers override earlier ones where applicable.

## Merge rules

### `patterns`
- Add a new pattern by using a new `name`
- Override a default pattern by reusing its `name`

### List fields
The following lists are merged and deduplicated across layers:
- `file_blocks`
- `search_blocks`
- `command_blocks`

## Pattern replacement modes

- Literal string (e.g. `"***FILTERED***"`)
- `"mask"` → same-length asterisks
- `"env_filter"` → environment style masking (`KEY=***FILTERED***`)

## Example user config (`~/.cc-filter/config.yaml`)

```yaml
patterns:
  - name: "company_api_key"
    regex: 'COMPANY_API_KEY=([a-zA-Z0-9]{32})'
    replacement: "***FILTERED***"

  - name: "openai_keys"
    regex: 'sk-[a-zA-Z0-9]{48}'
    replacement: "***COMPANY_FILTERED***"

file_blocks:
  - "*.private"
  - "company-secrets.json"

search_blocks:
  - "internal_token"

command_blocks:
  - "cat.*company"
```

## Example project config (`./config.yaml`)

```yaml
patterns:
  - name: "project_token"
    regex: 'PROJECT_TOKEN=([a-zA-Z0-9-_]{24})'
    replacement: "***PROJECT_FILTERED***"

file_blocks:
  - "deployment-secrets.yaml"
```

## Testing

```bash
echo "COMPANY_API_KEY=abc123def456ghi789jkl012mno345pqr" | cc-filter
```

See also: `configs/example-config.yaml`
