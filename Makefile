# Common tasks. Run `make help` to list them.
BIN := bin/toolips
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/sgnd/toolips/internal/cli.Version=$(VERSION)

STATICCHECK := v0.8.2
# Keep tool caches inside the repo so sandbox/CI can write them.
XDG_CACHE_HOME ?= $(CURDIR)/.cache
STATICCHECK_CACHE ?= $(XDG_CACHE_HOME)/staticcheck
export XDG_CACHE_HOME
export STATICCHECK_CACHE

PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64
export PLATFORMS

.PHONY: help build install test lint dist bins clean

help: ## Show this help
	@grep -E '^[a-z0-9-]+:.*##' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-15s %s\n", $$1, $$2}'

build: ## Build ./bin/toolips
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/toolips

install: ## Install toolips into $GOPATH/bin
	CGO_ENABLED=0 go install -ldflags "$(LDFLAGS)" ./cmd/toolips

test: ## Unit tests
	CGO_ENABLED=0 go test ./...

lint: ## gofmt + go vet + staticcheck
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)
	go vet ./...
	@mkdir -p "$(STATICCHECK_CACHE)"
	go run honnef.co/go/tools/cmd/staticcheck@$(STATICCHECK) ./...

GOBUILD = CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)"
OUT ?= bins

bins: ## Build release binaries into OUT/toolips_OS_ARCH/
	@for p in $(PLATFORMS); do \
	  os=$${p%/*}; arch=$${p#*/}; name=toolips_$${os}_$${arch}; \
	  exe=toolips; [ $$os = windows ] && exe=toolips.exe; \
	  echo "  $$name"; mkdir -p $(OUT)/$$name && \
	  GOOS=$$os GOARCH=$$arch $(GOBUILD) -o $(OUT)/$$name/$$exe ./cmd/toolips || exit 1; \
	done

dist: ## Build release archives + checksums into ./dist
	rm -rf dist && mkdir -p dist
	@for p in $(PLATFORMS); do \
	  os=$${p%/*}; arch=$${p#*/}; name=toolips_$${os}_$${arch}; \
	  exe=toolips; [ $$os = windows ] && exe=toolips.exe; \
	  mkdir -p dist/$$name && \
	  echo "  $$name" && GOOS=$$os GOARCH=$$arch $(GOBUILD) -o dist/$$name/$$exe ./cmd/toolips && \
	  cp LICENSE README.md dist/$$name/ && \
	  if [ $$os = windows ]; then (cd dist/$$name && zip -qX ../$$name.zip *); \
	  else tar -czf dist/$$name.tar.gz -C dist/$$name .; fi && \
	  rm -rf dist/$$name || exit 1; \
	done
	@expected=$$(set -- $(PLATFORMS); echo $$#); \
	got=$$(ls dist/*.tar.gz dist/*.zip 2>/dev/null | wc -l | tr -d ' '); \
	[ "$$got" = "$$expected" ] || { echo "dist: expected $$expected archives, got $$got"; ls -la dist; exit 1; }
	cd dist && (sha256sum *.tar.gz *.zip 2>/dev/null || shasum -a 256 *.tar.gz *.zip) > checksums.txt

clean: ## Remove build output
	rm -rf bin bins dist .cache .staticcheck-cache
