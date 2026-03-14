# cc-filter: Claude Code Sensitive Information Filter

```
 ██████╗ ██████╗     ███████╗██╗██╗  ████████╗███████╗██████╗
██╔════╝██╔════╝     ██╔════╝██║██║  ╚══██╔══╝██╔════╝██╔══██╗
██║     ██║          █████╗  ██║██║     ██║   █████╗  ██████╔╝
██║     ██║          ██╔══╝  ██║██║     ██║   ██╔══╝  ██╔══██╗
╚██████╗╚██████╔╝    ██║     ██║███████╗██║   ███████╗██║  ██║
 ╚═════╝ ╚═════╝     ╚═╝     ╚═╝╚══════╝╚═╝   ╚══════╝╚═╝  ╚═╝
```

>Claude: You are absolutely right, I can read everything from your `.env` file

>Claude: read `.env`

>Me: WTF! `.env` is on my denied list!

>Claude: Ah, I see the problem! I shouldn't have access to this file!

Claude Code, somewhere based on a true story.

## Overview

cc-filter adds a hard security layer in front of Claude Code hooks. It blocks sensitive file access, blocks risky shell/search commands, and redacts secrets from text.

It is designed to protect against bypasses (alternate paths, command tricks, indirect reads) that can slip past normal allow/deny patterns.

## What it protects

1. **Hard file blocks** (`.env`, key/cert files, secrets files)
2. **Command blocks** (e.g. commands trying to print secrets)
3. **Search blocks** (e.g. grep/find patterns targeting secrets)
4. **Prompt blocks** for `UserPromptSubmit` (exit code `2`, prompt never reaches Claude)
5. **Optional redaction** for selected source/config files

## Install

Download the latest release for your platform.

**macOS (Intel)**
```bash
curl -L -o cc-filter https://github.com/wissem/cc-filter/releases/latest/download/cc-filter-darwin-amd64
chmod +x cc-filter
sudo mv cc-filter /usr/local/bin/
```

**macOS (Apple Silicon)**
```bash
curl -L -o cc-filter https://github.com/wissem/cc-filter/releases/latest/download/cc-filter-darwin-arm64
chmod +x cc-filter
sudo mv cc-filter /usr/local/bin/
```

**Linux (x86_64)**
```bash
curl -L -o cc-filter https://github.com/wissem/cc-filter/releases/latest/download/cc-filter-linux-amd64
chmod +x cc-filter
sudo mv cc-filter /usr/local/bin/
```

**Windows (PowerShell)**
```powershell
Invoke-WebRequest -Uri "https://github.com/wissem/cc-filter/releases/latest/download/cc-filter-windows-amd64.exe" -OutFile "cc-filter.exe"
Move-Item cc-filter.exe C:\Windows\System32\
```

## Quick setup (Claude Code hooks)

Add this to Claude settings:
- global: `~/.claude/settings.json`
- project-specific: `.claude/settings.json`

```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": "*",
      "hooks": [{
        "type": "command",
        "command": "cc-filter"
      }]
    }],
    "UserPromptSubmit": [{
      "hooks": [{
        "type": "command",
        "command": "cc-filter"
      }]
    }],
    "SessionEnd": [{
      "hooks": [{
        "type": "command",
        "command": "cc-filter"
      }]
    }]
  }
}
```

## Standalone usage

```bash
echo "API_KEY=sk-1234567890abcdef" | cc-filter
# Output: API_KEY=***FILTERED***

# Filter API keys from files
cat config.txt | cc-filter

# Filter OpenAI keys
echo "My key is sk-1234567890abcdefghijklmnopqrstuvwxyz123456789012" | cc-filter
# Output: My key is ***************************************************
```

## Configuration basics

Configuration is layered (later overrides earlier):
1. `configs/default-rules.yaml` (built-in defaults)
2. `~/.cc-filter/config.yaml` (user-wide)
3. `./config.yaml` (project)

Merge behavior:
- `patterns`: add new names, or override defaults by reusing a name
- `file_blocks`, `search_blocks`, `command_blocks`: merged + deduplicated

Minimal user config example:

```yaml
patterns:
  - name: "company_api_key"
    regex: 'COMPANY_API_KEY=([a-zA-Z0-9]{32})'
    replacement: "***FILTERED***"

file_blocks:
  - "*.private"

search_blocks:
  - "internal_token"

command_blocks:
  - "cat.*company"
```

For a complete example, see `configs/example-config.yaml`.

## Docs

- [Configuration reference](docs/configuration.md)
- [Smart file redaction](docs/redaction.md)
- [Hook behavior and limitations](docs/hooks-and-limitations.md)
- [Default protected patterns/files](docs/default-rules.md)
- [Standalone + integrations](docs/integrations.md)
- [Logging + troubleshooting](docs/troubleshooting.md)

## License

MIT
