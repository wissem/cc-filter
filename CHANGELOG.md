# Changelog

All notable changes to cc-filter are documented in this file.

## [v0.0.6] - 2026-03-12

### Fixed
- Fix default rules not loading when running as a Claude Code hook (#6)
- Embed default rules into binary so all patterns are available regardless of working directory

### Added
- Bearer token filtering pattern
- CI pipeline with GitHub Actions for running tests on push and PRs

## [v0.0.5] - 2026-03-11

### Fixed
- Fix permissions prompt and deny list override (#5)

## [v0.0.4] - 2026-01-22

### Added
- Configurable file redaction via `redact_files` config
- UserPromptSubmit UX improvements: inline display and clipboard copy
- Unit tests and API limitation documentation
- Windows support documentation

### Fixed
- UserPromptSubmit hook not blocking content correctly

## [v0.0.3] - 2025-09-13

### Fixed
- Separate build and release jobs to avoid race condition in CI

## [v0.0.2] - 2025-09-13

### Changed
- Updated release workflow

## [v0.0.1] - 2025-09-13

### Added
- Initial release
- Stdin filtering with configurable regex patterns
- Claude Code hook support (PreToolUse, UserPromptSubmit, SessionEnd)
- Default rules for API keys, secret keys, access tokens, passwords, database URLs, JWT tokens, private keys, client secrets, auth tokens, OpenAI keys, Slack tokens, and environment variables
- File blocking for `.env`, `.pem`, `.key`, and other sensitive files
- Search and command blocking
- User config (`~/.cc-filter/config.yaml`) and project config (`./config.yaml`) support with merge strategy
