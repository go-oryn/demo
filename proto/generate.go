// Package proto holds the protobuf definitions of the gRPC API, and their generated Go code.
// The generated code is committed: regenerating it needs protoc (https://protobuf.dev/installation/).
package proto

//go:generate sh -c "protoc --plugin=protoc-gen-go=$(go tool -n protoc-gen-go) --plugin=protoc-gen-go-grpc=$(go tool -n protoc-gen-go-grpc) --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative tasks/v1/tasks.proto"
