BIN_DIR := $(CURDIR)/bin
LINUX_ARCH ?= amd64

.PHONY: build build-linux run clean

build:
	go build -o $(BIN_DIR)/build .

build-linux:
	GOOS=linux GOARCH=$(LINUX_ARCH) CGO_ENABLED=0 go build -o $(BIN_DIR)/build-linux-$(LINUX_ARCH) .

run: build
	$(BIN_DIR)/build $(CURDIR)

clean:
	rm -rf $(BIN_DIR)
