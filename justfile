# [Rust] compile packages and dependencies
build-rust:
    @echo "Building Rust user service..."
    cd rust-user-service && cargo build

# [Go] generate protobuf code
proto-go:
    @echo "Generating Go protobuf code..."
    mkdir -p go-chat-service/pkg/pb && \
    protoc --proto_path=./proto \
    --plugin=protoc-gen-go=`go env GOPATH`/bin/protoc-gen-go \
    --plugin=protoc-gen-go-grpc=`go env GOPATH`/bin/protoc-gen-go-grpc \
    --go_out=go-chat-service/pkg/pb --go_opt=paths=source_relative \
    --go-grpc_out=go-chat-service/pkg/pb --go-grpc_opt=paths=source_relative \
    chat.proto user.proto

# [Go] compile packages and dependencies
build-go: proto-go
    @echo "Building Go chat service..."
    cd go-chat-service && go build -o chat-server ./cmd/server

# build all services
build: build-rust build-go

# [Rust] remove object files and cached files
clean-rust:
    cd rust-user-service && cargo clean

# [Go] remove object files and cached files
clean-go:
    cd go-chat-service && go clean && rm -rf pkg/pb chat-server

# clean everything
clean: clean-rust clean-go

# [Rust] run user auth server
run-rust:
    @echo "Running Rust user service..."
    cd service && cargo run

# [Go] run chat service
run-go:
    @echo "Running Go chat service..."
    cd go-chat-service && ./chat-server
