.PHONY: build install clean run help

all: build

build:
	@echo "Building wydocli..."
	@go build -o wydocli ./cmd/wydocli
	@echo "Done: Build complete: ./wydocli"

install:
	@echo "Installing wydocli..."
	@go install ./cmd/wydocli
	@echo "Done: Installed to ~/go/bin/wydocli"

clean:
	@echo "Cleaning..."
	@rm -f wydocli
	@go clean
	@echo "Done: Clean complete"

run: build
	@./wydocli

help:
	@echo "wydocli - Documentation CLI Tool"
	@echo ""
	@echo "Available targets:"
	@echo "  make build    - Build the application (default)"
	@echo "  make install  - Install to ~/go/bin"
	@echo "  make clean    - Remove build artifacts"
	@echo "  make run      - Build and run the application"
	@echo "  make help     - Show this help message"
