.PHONY: build install test clean

PROTO_REF_DIR=$(CURDIR)/ratelproto
PROTO_REF_FILES=$(shell find "$(PROTO_REF_DIR)" -type f -name '*.proto')
compile-proto-ref:
	protoc  --go_out=$(PROTO_REF_DIR) --go_opt=paths=source_relative --proto_path=$(PROTO_REF_DIR) $(PROTO_REF_FILES)

# ============================================================================
# Build
# ============================================================================

build:
	go build -o $(CURDIR)/bin/ratel ./cmd/ratel
	go build -o $(CURDIR)/bin/protoc-gen-ratel ./cmd/protoc-gen-ratel

install:
	go install ./cmd/ratel
	go install ./cmd/protoc-gen-ratel

# ============================================================================
# Proto generation
# ============================================================================

RATELPROTO_DIR=$(CURDIR)/ratelproto
RATELPROTO_FILES=$(shell find "$(RATELPROTO_DIR)" -type f -name '*.proto')

.PHONY: proto
proto:
	protoc --go_out=$(RATELPROTO_DIR) --go_opt=paths=source_relative \
		--proto_path=$(RATELPROTO_DIR) \
		$(RATELPROTO_FILES)

# ============================================================================
# Example: Proto store
# ============================================================================

EXAMPLE_PROTO_DIR=$(CURDIR)/examples/proto
EXAMPLE_PROTO_FILES=$(shell find "$(EXAMPLE_PROTO_DIR)" -type f -name '*.proto')
EXAMPLE_PROTO_OUT=$(CURDIR)/examples/proto/storepb

.PHONY: example-proto-clean
example-proto-clean:
	rm -rf $(EXAMPLE_PROTO_OUT)
	mkdir -p $(EXAMPLE_PROTO_OUT)

.PHONY: example-proto
example-proto: build example-proto-clean
	protoc \
		--plugin=protoc-gen-ratel=$(CURDIR)/bin/protoc-gen-ratel \
		--go_out=$(CURDIR) \
		--go_opt=paths=source_relative \
		--ratel_out=$(CURDIR) \
		--ratel_opt=paths=source_relative \
		--proto_path=$(CURDIR) \
		$(EXAMPLE_PROTO_FILES)

# ============================================================================
# Example: Store (hand-written models)
# ============================================================================

.PHONY: example-store-test
example-store-test:
	cd examples/store && go test -v -timeout 5m

# ============================================================================
# Tests
# ============================================================================

.PHONY: test
test:
	go test -v ./...

.PHONY: test-integration
test-integration:
	go test -v -timeout 10m ./examples/store/...

# ============================================================================
# Clean
# ============================================================================

clean:
	rm -rf $(CURDIR)/bin
	rm -rf $(EXAMPLE_PROTO_OUT)

# ============================================================================
# CLI commands demo
# ============================================================================

.PHONY: demo-generate
demo-generate: install
	ratel generate -i examples/store/gold.sql -o /tmp/ratel-demo/models -p models
	@echo "Generated models in /tmp/ratel-demo/models"
	@ls -la /tmp/ratel-demo/models/

.PHONY: demo-schema
demo-schema: install
	ratel schema -i examples/store/models -o /tmp/ratel-demo/schema.sql
	@echo "Generated schema:"
	@cat /tmp/ratel-demo/schema.sql


branch=main
.PHONY: revision
revision: # Создание тега
	@if [ -e $(tag) ]; then \
		echo "error: Specify version 'tag='"; \
		exit 1; \
	fi
	git tag -d ${tag} || true
	git push --delete origin ${tag} || true
	git tag $(tag)
	git push origin $(tag)