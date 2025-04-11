run:
	@cd src/cmd/main && go run main.go

build:
	@echo "Building chat service..."
	@mkdir -p bin
	@cd src && go build -o ../bin/chat ./cmd/main
	@echo "Build complete. Binary available at bin/chat"