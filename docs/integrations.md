# Standalone usage and integrations

cc-filter works as a stdin/stdout filter.

## Standalone usage examples

```bash
# Basic key filtering
echo "API_KEY=sk-1234567890abcdef" | cc-filter
# Expected output:
# API_KEY=***FILTERED***

# OpenAI-style key masking
echo "My key is sk-1234567890abcdefghijklmnopqrstuvwxyz123456789012" | cc-filter
# Expected output:
# My key is ***************************************************

# Filter a file through stdin
cat config.txt | cc-filter
# Example expected output:
# DB_URL=***FILTERED***
# PASSWORD=***FILTERED***
```

## Integration model

Any coding agent/tool can integrate cc-filter if it supports one of:
- command hooks
- pre-processing commands
- pipe-based filtering

cc-filter auto-detects Claude-style JSON hook payloads and falls back to plain text filtering when input is not hook JSON.
