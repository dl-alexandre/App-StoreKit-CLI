# App Store Server CLI (iap)

`iap` is a Go-based CLI for the App Store Server API.

## Install (development)

```bash
go build -o iap ./cmd/iap
```

## Install (Homebrew)

```bash
brew tap dl-alexandre/homebrew-tap
brew install iap
```

## Configuration

Default config path:

```
~/.config/iap/config.yaml
```

Initialize config interactively:

```bash
iap config init
```

Validate config:

```bash
iap config validate
```

## 5-minute quickstart

```bash
iap config init
iap config validate
make test-smoke
```

Environment variables:

- `IAP_ISSUER_ID`
- `IAP_KEY_ID`
- `IAP_BUNDLE_ID`
- `IAP_PRIVATE_KEY_PATH`
- `IAP_PRIVATE_KEY`
- `IAP_ENV` (`sandbox`, `production`, or `local-testing`)
- `IAP_MAX_RETRIES`
- `IAP_RETRY_BACKOFF_MS`

Precedence:

1. CLI flags
2. Environment variables
3. Config file

## Usage

```bash
iap transaction get --transaction-id <id>
iap transaction history --transaction-id <id> --version v2
iap transaction app --transaction-id <id>
iap transaction app-account-token --original-transaction-id <id> --body request.json
iap notification history --body request.json
iap notification test
iap notification test-status --test-notification-token <token>
iap refund history --transaction-id <id>
iap subscription status --transaction-id <id>
iap subscription extend --original-transaction-id <id> --body request.json
iap subscription extend-mass --body request.json
iap subscription extend-status --product-id <id> --request-id <uuid>
iap order lookup --order-id <id>
iap consumption send --transaction-id <id> --body request.json
iap messaging image list
iap messaging image upload --image-id <uuid> --file image.png
iap messaging image delete --image-id <uuid>
iap messaging message list
iap messaging message upload --message-id <uuid> --body request.json
iap messaging message delete --message-id <uuid>
iap messaging default configure --product-id <id> --locale <locale> --body request.json
iap messaging default delete --product-id <id> --locale <locale>
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
iap transaction get --transaction-id <id>
iap transaction history --transaction-id <id> --version v2
iap transaction app --transaction-id <id>
```

Notifications:

```bash
iap notification test
iap notification test-status --test-notification-token <token>
iap notification history --body request.json
```

Subscriptions:

```bash
iap subscription status --transaction-id <id>
iap subscription extend --original-transaction-id <id> --body request.json
iap subscription extend-status --product-id <id> --request-id <uuid>
```

Refunds + orders:

```bash
iap refund history --transaction-id <id>
iap order lookup --order-id <id>
```

Messaging:

```bash
iap messaging image list
iap messaging message upload --message-id <uuid> --body request.json
iap messaging default configure --product-id <id> --locale <locale> --body request.json
```

Consumption:

```bash
iap consumption send --transaction-id <id> --body request.json
```

## Smoke test

```bash
make test-smoke
```

## Shell completion

```bash
iap completion bash > /usr/local/etc/bash_completion.d/iap
iap completion zsh > /usr/local/share/zsh/site-functions/_iap
iap completion fish > ~/.config/fish/completions/iap.fish
```

## Notes

This CLI expects raw JSON request bodies for POST endpoints. Use `--body -` to read from stdin.
