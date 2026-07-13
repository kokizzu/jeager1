GO ?= go
GOVULNCHECK ?= govulncheck
CMD ?=
.PHONY: test verify-dependency-security run vulncheck

test:
	$(GO) test ./...

verify-dependency-security:
	bash ./scripts/verify-dependency-security.sh

vulncheck:
	$(GOVULNCHECK) ./...

run:
	@test -n "$(CMD)" || (echo "usage: make run CMD='go test ./...'" >&2; exit 2)
	$(CMD)
