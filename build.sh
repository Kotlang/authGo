rm -Rf generated
mkdir generated
cd auth-model
protoc --go_out=../generated --go_opt=paths=source_relative \
    --go-grpc_out=../generated --go-grpc_opt=paths=source_relative \
    *.proto
cd ..
go build -mod=mod -o build/ .