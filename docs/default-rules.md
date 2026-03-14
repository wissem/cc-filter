# Default protected patterns and files

Built-in defaults live in:
- `configs/default-rules.yaml`

## Typical sensitive patterns

- API keys (`api_key`, `api-key`)
- secret keys
- access tokens
- passwords
- database URLs
- JWTs
- private keys
- client/auth tokens
- OpenAI keys (`sk-...`)
- Slack bot tokens (`xoxb-...`)
- env-style secrets (`KEY=value`)

## Typical blocked file patterns

- `.env` files
- key/cert files (`.key`, `.pem`, `.p12`, `.pfx`)
- common secret json names (`config.json`, `secrets.json`, `credentials.json`, `auth.json`, `keys.json`)

For exact and current rules, always check `configs/default-rules.yaml`.
