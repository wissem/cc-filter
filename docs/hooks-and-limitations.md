# Hook behavior and limitations

## Hook roles

- `PreToolUse`: blocks/filters tool input (read, search, command patterns)
- `UserPromptSubmit`: scans prompts before Claude sees them
- `SessionEnd`: cleanup (e.g. temp redacted files)

## Exit codes

- `0`: success, pass through
- `1`: error
- `2`: blocked

`Exit code 2` is important for `UserPromptSubmit`: it blocks submission and prevents the original prompt from entering model context.

## Current Claude Code hook API limits

### File reads
Seamless path rewrite is limited in current hook behavior. Practical flow is:
1. deny original read
2. create redacted copy
3. tell Claude to read redacted path

### User prompts
Prompt text cannot be rewritten in place by hook API. Available behaviors are allow/block (+ context). So cc-filter blocks and shows redacted output for easy re-paste.

These constraints come from hook API behavior, not cc-filter filtering logic.
