# App StoreKit CLI (ask)

[![Go Report Card](https://goreportcard.com/badge/github.com/dl-alexandre/App-StoreKit-CLI)](https://goreportcard.com/report/github.com/dl-alexandre/App-StoreKit-CLI)

`ask` is a Go-based CLI for the App Store Server API and External Purchase Server API.

## Install (development)

```bash
go build -o ask ./cmd/ask
```

## Install (Homebrew)

```bash
brew tap dl-alexandre/homebrew-tap
brew install ask
```

## Configuration

Default config path:

```
~/.config/ask/config.yaml
```

Initialize config interactively:

```bash
ask config init
```

Validate config:

```bash
ask config validate
```

## 5-minute quickstart

```bash
ask config init
ask config validate
make test-smoke
```

Environment variables:

- `ASK_ISSUER_ID`
- `ASK_KEY_ID`
- `ASK_BUNDLE_ID`
- `ASK_PRIVATE_KEY_PATH`
- `ASK_PRIVATE_KEY`
- `ASK_ENV` (`sandbox`, `production`, or `local-testing`)
- `ASK_MAX_RETRIES`
- `ASK_RETRY_BACKOFF_MS`

Precedence:

1. CLI flags
2. Environment variables
3. Config file

## Usage

```bash
ask transaction get --transaction-id <id>
ask transaction history --transaction-id <id> --version v2
ask transaction app --transaction-id <id>
ask transaction app-account-token --original-transaction-id <id> --body request.json
ask notification history --body request.json
ask notification test
ask notification test-status --test-notification-token <token>
ask refund history --transaction-id <id>
ask subscription status --transaction-id <id>
ask subscription extend --original-transaction-id <id> --body request.json
ask subscription extend-mass --body request.json
ask subscription extend-status --product-id <id> --request-id <uuid>
ask order lookup --order-id <id>
ask consumption send --transaction-id <id> --body request.json
ask external send --body report.json
ask external get --request-id <uuid>
ask messaging image list
ask messaging image upload --image-id <uuid> --file image.png
ask messaging image delete --image-id <uuid>
ask messaging message list
ask messaging message upload --message-id <uuid> --body request.json
ask messaging message delete --message-id <uuid>
ask messaging default configure --product-id <id> --locale <locale> --body request.json
ask messaging default delete --product-id <id> --locale <locale>
```

Common flags:

- `--format json|table|raw`
- `--jq '<expression>'`
- `--query key=value` (repeatable)
- `--table-columns col1,col2` (table output only)
- `--debug`

## Examples

Transactions:

```bash
ask transaction get --transaction-id <id>
ask transaction history --transaction-id <id> --version v2
ask transaction app --transaction-id <id>
```

Notifications:

```bash
ask notification test
ask notification test-status --test-notification-token <token>
ask notification history --body request.json
```

Subscriptions:

```bash
ask subscription status --transaction-id <id>
ask subscription extend --original-transaction-id <id> --body request.json
ask subscription extend-status --product-id <id> --request-id <uuid>
```

Refunds + orders:

```bash
ask refund history --transaction-id <id>
ask order lookup --order-id <id>
```

Messaging:

```bash
ask messaging image list
ask messaging message upload --message-id <uuid> --body request.json
ask messaging default configure --product-id <id> --locale <locale> --body request.json
```

Consumption:

```bash
ask consumption send --transaction-id <id> --body request.json
```

External purchases:

```bash
ask external send --body report.json
ask external get --request-id <uuid>
```

## Smoke test

```bash
make test-smoke
```

## Shell completion

```bash
ask completion bash > /usr/local/etc/bash_completion.d/ask
ask completion zsh > /usr/local/share/zsh/site-functions/_ask
ask completion fish > ~/.config/fish/completions/ask.fish
```

## Notes

This CLI expects raw JSON request bodies for POST endpoints. Use `--body -` to read from stdin.
