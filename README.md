dist013
ehe6gqRsS2Fk
10.35.168.23
lester
50051

dist014
KRZ65kfAEmpB
10.35.168.24
Michael
50052

dist015
aNASDGkYnQ8F
10.35.168.25
Trevor
50053

dist016
jrKU59Umn2TW
10.35.168.26
Franklin
50054


protoc --go_out=. --go-grpc_out=. main.proto
protoc --go_out=./lester --go-grpc_out=./lester ./proto/heist.proto && protoc --go_out=./michael --go-grpc_out=./michael ./proto/heist.proto


