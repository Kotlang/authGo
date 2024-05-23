rm -Rf generated
rm wire_gen.go

mkdir -p generated/auth
cd auth-model
protoc --go_out=../generated/auth --go_opt=paths=source_relative \
    --go-grpc_out=../generated/auth --go-grpc_opt=paths=source_relative \
    *.proto
cd ..

mkdir -p generated/notification
cd notification-model
protoc --go_out=../generated/notification --go_opt=paths=source_relative \
    --go-grpc_out=../generated/notification --go-grpc_opt=paths=source_relative \
    *.proto
cd ..

go install github.com/google/wire/cmd/wire@latest
wire

go build -mod=mod -o build/ .
