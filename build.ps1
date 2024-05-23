Remove-Item generated -Recurse
Remove-Item wire_gen.go

New-Item -Path . -Name "generated\auth" -ItemType "directory"
New-Item -Path . -Name "generated\notification" -ItemType "directory"

cd auth-model
protoc --go_out=../generated/auth --go_opt=paths=source_relative `
    --go-grpc_out=../generated/auth --go-grpc_opt=paths=source_relative `
    *.proto
cd ..

cd notification-model
protoc --go_out=../generated/notification --go_opt=paths=source_relative `
    --go-grpc_out=../generated/notification --go-grpc_opt=paths=source_relative `
    *.proto
cd ..

go install github.com/google/wire/cmd/wire@latest
wire.exe

go build -mod=mod -o build/ .