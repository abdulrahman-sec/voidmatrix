# Makefile for voidmatrix

BINARY_NAME=voidmatrix
PREFIX?=/usr/local
INSTALL_PATH=$(PREFIX)/bin

.PHONY: all build install uninstall clean

all: build

build:
	go build -o $(BINARY_NAME) .

install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	@mkdir -p $(INSTALL_PATH)
	@cp $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installation successful! Run '$(BINARY_NAME)' from your terminal."

uninstall:
	@echo "Removing $(BINARY_NAME) from $(INSTALL_PATH)..."
	@rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Uninstallation successful."

clean:
	rm -f $(BINARY_NAME)
