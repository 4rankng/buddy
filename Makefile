SHELL := /bin/bash
BINARY := oncallmy
CMD_PKG := .
BIN_DIR := $(HOME)/bin

.PHONY: all build deploy clean

all: build

build:
	go build -o $(BINARY) $(CMD_PKG)

deploy: build
	mkdir -p "$(BIN_DIR)"
	mv -f "$(BINARY)" "$(BIN_DIR)/$(BINARY)"
	@echo "Deployed to $(BIN_DIR)/$(BINARY)"

clean:
	rm -f "$(BINARY)"
