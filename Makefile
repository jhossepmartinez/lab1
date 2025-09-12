.PHONY: proto lester michael franklin trevor

proto:
	protoc --go_out=./lester --go-grpc_out=./lester ./proto/heist.proto && \
	protoc --go_out=./michael --go-grpc_out=./michael ./proto/heist.proto && \
	protoc --go_out=./franklin --go-grpc_out=./franklin ./proto/heist.proto && \
	protoc --go_out=./trevor --go-grpc_out=./trevor ./proto/heist.proto

lester:
	cd ./lester && go run ./main.go

michael:
	cd ./michael/ && go run ./main.go

franklin:
	cd ./franklin/ && go run ./main.go

trevor:
	cd ./trevor/ && go run ./main.go


