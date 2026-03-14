# Logging and troubleshooting

cc-filter logs activity to:
- `~/.cc-filter/filter.log`

## What is logged

- invocation timestamp
- input type (hook JSON vs plain text)
- input/output size
- processing duration

## Useful checks

```bash
# Follow live log output
tail -f ~/.cc-filter/filter.log

# Confirm binary path
which cc-filter

# Confirm user config exists (optional)
ls ~/.cc-filter/config.yaml
```

If hooks seem inactive, verify Claude settings path and JSON structure in your `settings.json`.
