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
cc-filter is a security tool that enforces strict access controls for Claude Code by intercepting and filtering data before it reaches the AI. 

Unlike Claude Code's built-in permissions and CLAUDE.md rules which can be bypassed through various methods (alternative file paths, indirect commands, etc.), cc-filter provides an additional security layer that cannot be circumvented. It prevents Claude from reading sensitive files like `.env`, executing commands that expose credentials, or accessing API keys regardless of how the request is formatted. 

The tool comes with a comprehensive default configuration for common secrets and allows full customization through editable configuration files. Additionally, cc-filter provides a more powerful and flexible filtering system than basic pattern matching - supporting regex patterns, multiple replacement strategies, file-type aware filtering, and command-line argument analysis for complete protection coverage.

## Installation

Download the latest release for your platform:


**macOS (Intel):**
```bash
curl -L -o cc-filter https://github.com/wissem/cc-filter/releases/latest/download/cc-filter-darwin-amd64
chmod +x cc-filter
sudo mv cc-filter /usr/local/bin/
```


**macOS (Apple Silicon):**
```bash
curl -L -o cc-filter https://github.com/wissem/cc-filter/releases/latest/download/cc-filter-darwin-arm64
chmod +x cc-filter
sudo mv cc-filter /usr/local/bin/
```

**Linux (x86_64):**
```bash
curl -L -o cc-filter https://github.com/wissem/cc-filter/releases/latest/download/cc-filter-linux-amd64
chmod +x cc-filter
sudo mv cc-filter /usr/local/bin/
```

**Windows (PowerShell):**
```powershell
# Download the binary
Invoke-WebRequest -Uri "https://github.com/wissem/cc-filter/releases/latest/download/cc-filter-windows-amd64.exe" -OutFile "cc-filter.exe"

# Move to a directory in your PATH (e.g., C:\Windows\System32 or create a custom bin folder)
Move-Item cc-filter.exe C:\Windows\System32\
```

Alternatively, you can download `cc-filter-windows-amd64.exe` from the [releases page](https://github.com/wissem/cc-filter/releases/latest) and place it in a directory that's in your PATH.

## How It Works

cc-filter intercepts data at multiple points in the Claude Code workflow:

```
┌─────────────────────────────────────────────────────────────────┐
│                        USER PROMPT FLOW                         │
├─────────────────────────────────────────────────────────────────┤
│  User types prompt → cc-filter scans → Secrets detected?        │
│                                          │                      │
│                                    NO ───┴─── YES               │
│                                    │           │                │
│                                    ▼           ▼                │
│                              Pass through   BLOCK (exit 2)      │
│                              (exit 0)       + create redacted   │
│                                             file for reference  │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                        FILE READ FLOW                           │
├─────────────────────────────────────────────────────────────────┤
│  Claude reads file → cc-filter checks → Blocked file type?      │
│                                          │                      │
│                                    NO ───┴─── YES               │
│                                    │           │                │
│                                    ▼           ▼                │
│                           Scan for secrets   DENY access        │
│                                    │                            │
│                              NO ───┴─── YES                     │
│                              │           │                      │
│                              ▼           ▼                      │
│                         Allow read   Redirect to                │
│                                      redacted copy              │
└─────────────────────────────────────────────────────────────────┘
```

## Exit Codes

cc-filter uses exit codes to communicate with Claude Code hooks:

| Exit Code | Meaning | Effect |
|-----------|---------|--------|
| 0 | Success | Content passed through unchanged |
| 1 | Error | Initialization or processing failed |
| 2 | **Blocked** | Content rejected, prompt erased from context |

> **Important:** Exit code 2 is crucial for `UserPromptSubmit` hooks. It signals that the prompt should be **blocked AND erased** from the conversation context, preventing the AI from ever seeing the sensitive data.

## Usage with Claude Code Hooks

### 1. Create a Claude Code configuration

Create or update your Claude Code configuration file (usually `~/.claude/settings.json` or project-specific):

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

**Hook explanations:**
- **PreToolUse**: Intercepts tool calls (Read, Bash, Grep, Glob) to block or redact sensitive file access
- **UserPromptSubmit**: Scans user prompts for secrets before they reach Claude (blocks with exit code 2)
- **SessionEnd**: Cleans up temporary redacted files when the session ends

### 2. Project-specific usage

For project-specific filtering, create `.claude/settings.json` in your project root:

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
## Configuration

cc-filter uses a flexible configuration system that allows you to extend or customize filtering rules without replacing the built-in defaults.

### Configuration Files (loaded in order)

1. **Default Rules** - Built-in filtering patterns (`configs/default-rules.yaml`)
2. **User Configuration** - Your global customizations (`~/.cc-filter/config.yaml`)
3. **Project Configuration** - Project-specific rules (`config.yaml` in current directory)

### How Configuration Merging Works

**For Patterns:**
- **Extension**: Add new patterns by giving them unique names
- **Override**: Replace default patterns by using the same `name`

**For Lists** (file_blocks, search_blocks, command_blocks):
- **Extension**: All items from all configs are combined (duplicates removed)

### Creating Your Configuration

**No copying required!** Just specify what you want to add or change.

#### Example User Config (`~/.cc-filter/config.yaml`):

```yaml
patterns:
  # Add a new custom pattern
  - name: "company_api_key"
    regex: 'COMPANY_API_KEY=([a-zA-Z0-9]{32})'
    replacement: "***FILTERED***"

  # Override the default openai_keys pattern
  - name: "openai_keys"
    regex: 'sk-[a-zA-Z0-9]{48}'
    replacement: "***COMPANY_FILTERED***"

# Add additional file patterns to block
file_blocks:
  - "*.private"
  - "company-secrets.json"

# Add additional search terms to block
search_blocks:
  - "internal_token"
  - "company_secret"

# Add additional command patterns to block
command_blocks:
  - "cat.*company"
  - "grep.*internal"
```

#### Example Project Config (`config.yaml`):

```yaml
patterns:
  # Project-specific API pattern
  - name: "project_token"
    regex: 'PROJECT_TOKEN=([a-zA-Z0-9-_]{24})'
    replacement: "***PROJECT_FILTERED***"

file_blocks:
  - "deployment-secrets.yaml"
  - "*.production"
```

### Pattern Replacement Types

- `"***FILTERED***"` - Replace with literal text
- `"mask"` - Replace with asterisks (`*`) matching original length
- `"env_filter"` - For environment variables: `KEY=***FILTERED***`

### Testing Your Configuration

```bash
# Test custom patterns
echo "COMPANY_API_KEY=abc123def456ghi789jkl012mno345pqr" | ./cc-filter

# Test file blocking (when used with Claude Code hooks)
# Attempts to read company-secrets.json will be blocked

# Test search blocking (when used with Claude Code hooks)
# Searches containing "internal_token" will be blocked
```

### Configuration Examples

See `configs/example-config.yaml` for a complete example showing all available options.

## Filtered Patterns

- API keys (api_key, api-key)
- Secret keys (secret_key, secret-key)
- Access tokens (access_token, access-token)
- Passwords
- Database URLs
- JWT tokens
- Private keys
- Client secrets
- Auth tokens
- OpenAI API keys (sk-...)
- Slack bot tokens (xoxb-...)
- Environment variables (KEY=value format)

## File Types Filtered

- .env files
- .key, .pem, .p12, .pfx files
- config.json, secrets.json, credentials.json
- auth.json, keys.json

The tool preserves the structure of your content while replacing sensitive values with `***FILTERED***` or asterisks.

## UserPromptSubmit Protection

When you use the `UserPromptSubmit` hook, cc-filter scans your prompts **before** they reach Claude:

### What happens when secrets are detected:

1. **Prompt is blocked** - The submission is rejected with exit code 2
2. **Prompt is erased** - Claude never sees the sensitive content
3. **Redacted copy created** - A sanitized version is saved to `/tmp/claude/redacted/user_input_*.txt`
4. **User notified** - Error message tells you where to find the redacted version

### Example scenario:

```
You: Here's my config: API_KEY=sk-1234567890abcdef...

cc-filter: BLOCKED: Sensitive content detected in your prompt.

           A redacted version has been saved to:
             /tmp/claude/redacted/user_input_a1b2c3d4.txt

           Please reference that file.
```

This prevents accidental exposure of secrets when copy-pasting code or configurations into Claude.

## Smart File Redaction

Instead of simply blocking all potentially sensitive files, cc-filter uses **smart redaction** for source code files:

### How it works:

1. When Claude tries to read a code file, cc-filter scans it for secrets
2. If secrets are found, a **redacted copy** is created in `/tmp/claude/redacted/`
3. Claude is redirected to read the redacted version instead
4. The redacted file includes a header noting it's been filtered

### File types that get smart redaction:

| Category | Extensions |
|----------|------------|
| Swift/Obj-C | `.swift`, `.m`, `.h` |
| JVM | `.kt`, `.java` |
| Scripting | `.py`, `.rb` |
| Systems | `.go`, `.rs` |
| Web | `.js`, `.ts`, `.tsx`, `.jsx` |
| Config | `.json`, `.yaml`, `.yml`, `.toml`, `.plist`, `.xcconfig` |

Files with names containing `config`, `environment`, `settings`, `secrets`, or `apikeys` are also scanned.

### Redacted file format:

```
# ***FILTERED*** REDACTED VERSION - Some sensitive values have been masked
# Original: /path/to/your/config.swift

let apiKey = "***FILTERED***"
let endpoint = "https://api.example.com"
```

### Cleanup

Redacted files are stored in `/tmp/claude/redacted/` and are automatically cleaned up when:
- The `SessionEnd` hook fires (end of Claude Code session)
- You manually delete the directory

## Standalone Usage

cc-filter accepts stdin input and can be adapted for use with any coding agent or tool that supports command-line filtering:

```bash
echo "API_KEY=sk-1234567890abcdef" | cc-filter
# Output: API_KEY=***FILTERED***

# Filter API keys from files
cat config.txt | cc-filter

# Filter OpenAI keys
echo "My key is sk-1234567890abcdefghijklmnopqrstuvwxyz123456789012" | cc-filter
# Output: My key is ***************************************************
```

### Integration with Other AI Coding Agents

Since cc-filter processes stdin/stdout, it can be integrated with any coding agent that supports:
- Pre-processing hooks
- Command-line filters
- Pipe-based text processing

The tool auto-detects JSON hook format but falls back to plain text filtering, making it compatible with various agent architectures beyond Claude Code. For different AI tools, you may need to add custom hook formats by extending the processors in the `internal/hooks/` directory.

## Logging

cc-filter automatically logs its activity to help you monitor when it's being invoked:

- **Log location**: `~/.cc-filter/filter.log`
- **Log format**: Standard timestamp with invocation details
- **Information logged**:
  - Invocation timestamp
  - Input type (JSON hook input or plain text)  
  - Content size (input/output bytes)
  - Processing duration

### Example log entries:
```
2025/09/09 10:30:45 cc-filter invoked at 2025-09-09T10:30:45-07:00
2025/09/09 10:30:45 Processing completed - Type: JSON hook input, Input: 1024 bytes, Output: 987 bytes, Duration: 2.1ms
```

### View recent activity:
```bash
tail -f ~/.cc-filter/filter.log
```

