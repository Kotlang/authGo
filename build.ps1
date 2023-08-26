Remove-Item generated -Recurse
New-Item -Path . -Name "generated" -ItemType "directory"

cd auth-model
protoc --go_out=../generated --go_opt=paths=source_relative `
    --go-grpc_out=../generated --go-grpc_opt=paths=source_relative `
    *.proto
cd ..

cd notification-model
protoc --go_out=../generated --go_opt=paths=source_relative `
    --go-grpc_out=../generated --go-grpc_opt=paths=source_relative `
    *.proto
cd ..

go build -mod=mod -o build/ .