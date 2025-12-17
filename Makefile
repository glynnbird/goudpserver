# ---- Config ----
APP_NAME := goudpserver
CMD_DIR  := .
BIN_DIR  := ./bin
BIN      := $(BIN_DIR)/$(APP_NAME)

GO       := go
GOFLAGS  := -trimpath
LDFLAGS  := -s -w

# ---- Phony targets ----
.PHONY: all build run test test-race fmt vet lint clean

all: build

# ---- Build ----
build:
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN) $(CMD_DIR)

# ---- Run ----
run:
	$(GO) run $(CMD_DIR)

# ---- Test ----
test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

# ---- Code quality ----
fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

lint:
	golangci-lint run

# ---- Clean ----
clean:
	rm -rf $(BIN_DIR)
