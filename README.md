## How to install
1. Install go.
2. Make sure GOPATH is set permanently. I did this by editng ~/.profile
    * `export GOPATH={path/to/go}`
3. Install protoc with `apt install -y protobuf-compiler` or follow:
    * https://grpc.io/docs/protoc-installation/
4. Call `go get -u github.com/golang/protobuf/protoc-gen-go` from anywhere.
