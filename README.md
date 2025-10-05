# cc-filter: Claude Code Sensitive Information Filter

```
 ██████╗ ██████╗     ███████╗██╗██╗  ████████╗███████╗██████╗
██╔════╝██╔════╝     ██╔════╝██║██║  ╚══██╔══╝██╔════╝██╔══██╗
██║     ██║          █████╗  ██║██║     ██║   █████╗  ██████╔╝
██║     ██║          ██╔══╝  ██║██║     ██║   ██╔══╝  ██╔══██╗
╚██████╗╚██████╔╝    ██║     ██║███████╗██║   ███████╗██║  ██║
 ╚═════╝ ╚═════╝     ╚═╝     ╚═╝╚══════╝╚═╝   ╚══════╝╚═╝  ╚═╝
```

>You are absolutely right, I can read everything from your `.env` file . Ah, I >see the problem. I shouldn't have access to this file!

Claude Code, somewhere

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
    }]
  }
}
```

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
    }]
  }
}
```

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

