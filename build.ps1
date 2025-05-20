Remove-Item generated -Recurse

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

go build -mod=mod -o build/ .