Step 2: Generate Go Code

Install the necessary tools if you haven't already:
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

Make sure the protoc compiler is installed on your system. Then, from the root of your project, run this command:
Shell

protoc --proto_path=proto \
  --go_out=gen/go --go_opt=paths=source_relative \
  --go-grpc_out=gen/go --go-grpc_opt=paths=source_relative \
  proto/ariand/v1/*.proto

This will create a gen/go/ariand/v1 directory containing the generated Go files (core.pb.go, services.pb.go, services_grpc.pb.go).