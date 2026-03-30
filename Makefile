# ABOUTME: Build, test, install, and release targets for the sanitize CLI tool.
# Standard entry points so any user can run make build/test/install without
# needing to know Go toolchain details.

BINARY := sanitize
BUILD_DIR := .
GO := go
REPO := tigger04/oed-sanitize
TAP_REPO := tigger04/homebrew-tap
CURRENT_VERSION := $(shell cat VERSION | tr -d '[:space:]')
VERSION ?= $(CURRENT_VERSION)
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: build test test-one-off install uninstall clean release sync

build:
	$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/sanitize/

test: build
	$(GO) test ./pkg/spelling/ -v
	$(GO) test ./tests/regression/ -v

test-one-off:
ifdef ISSUE
	$(GO) test ./tests/one_off/ -v -run "$(ISSUE)"
else
	$(GO) test ./tests/one_off/ -v
endif

install: build
	@mkdir -p ~/bin
	cp $(BUILD_DIR)/$(BINARY) ~/bin/$(BINARY)
	@echo "Installed $(BINARY) $(VERSION) to ~/bin/$(BINARY)"

uninstall:
	rm -f ~/bin/$(BINARY)
	@echo "Removed $(BINARY) from ~/bin"

clean:
	rm -f $(BUILD_DIR)/$(BINARY)

# Release workflow:
#   make release                    — increment minor version (0.1.0 → 0.2.0)
#   make release VERSION=1.0.0     — set explicit version
#   SKIP_TESTS=1 make release      — skip regression tests
release:
ifndef SKIP_TESTS
	@echo "Running regression tests..."
	$(MAKE) test
endif
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Error: working tree is dirty. Commit or stash changes first."; \
		exit 1; \
	fi
	@./scripts/release.sh "$(VERSION)" "$(CURRENT_VERSION)" "$(REPO)" "$(TAP_REPO)" "$(BINARY)"

sync:
	git add --all
	git commit -m "chore: sync" || true
	git pull --rebase
	git push
