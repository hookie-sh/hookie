.PHONY: proto-relay proto-cli proto

# Generate protobuf code for relay service
proto-relay:
	cd backend/relay && \
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/relay.proto

# Generate protobuf code for CLI
proto-cli:
	cd cli && \
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/relay.proto

# Generate protobuf code for both
proto: proto-relay proto-cli

# Build relay service
build-relay:
	cd backend/relay && go build -o ../../bin/relay main.go

# Build GUI (Vite)
build-gui:
	pnpm --filter gui build

# Build CLI (development by default, outputs to bin/hookie)
build-cli: build-cli-dev

# Build CLI for development (includes embedded GUI)
build-cli-dev: build-gui
	rm -rf cli/internal/gui/dist
	cp -r apps/gui/dist cli/internal/gui/dist
	cd cli && go build -tags dev -o ../bin/hookie main.go

# Build CLI for production (includes embedded GUI)
build-cli-prod: build-gui
	rm -rf cli/internal/gui/dist
	cp -r apps/gui/dist cli/internal/gui/dist
	cd cli && go build -o ../bin/hookie main.go

# Build both
build: build-relay build-cli

# Install dependencies for relay
deps-relay:
	cd backend/relay && go mod download && go mod tidy

# Install dependencies for CLI
deps-cli:
	cd cli && go mod download && go mod tidy

# Install dependencies for both
deps: deps-relay deps-cli

