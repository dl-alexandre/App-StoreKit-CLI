.PHONY: test-smoke
.PHONY: release-snapshot

test-smoke:
	go run ./cmd/ask config validate
	go run ./cmd/ask notification test

release-snapshot:
	goreleaser release --snapshot --clean
