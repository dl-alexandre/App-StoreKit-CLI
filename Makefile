.PHONY: test-smoke
.PHONY: release-snapshot

test-smoke:
	go run ./cmd/iap config validate
	go run ./cmd/iap notification test

release-snapshot:
	goreleaser release --snapshot --clean
