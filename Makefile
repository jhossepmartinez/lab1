.PHONY: proto lester michael franklin trevor docker-build docker-run-lester docker-run-michael docker-run-franklin docker-run-trevor

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

# rabbitmq-setup:
# 	@echo "Stopping and removing existing RabbitMQ container..."
# 	docker stop rabbitmq 2>/dev/null || true
# 	docker rm rabbitmq 2>/dev/null || true
# 	@echo "Starting RabbitMQ container..."
# 	docker run -d --name rabbitmq -p 5673:5672 -p 15673:15672 rabbitmq:4-management
# 	@echo "Waiting for RabbitMQ to start..."
# 	sleep 10
# 	@echo "Creating admin user..."
# 	docker exec rabbitmq rabbitmqctl wait /var/lib/rabbitmq/mnesia/rabbit@$(docker exec rabbitmq hostname).pid
# 	docker exec rabbitmq rabbitmqctl add_user admin admin
# 	docker exec rabbitmq rabbitmqctl set_user_tags admin administrator
# 	docker exec rabbitmq rabbitmqctl set_permissions -p / admin ".*" ".*" ".*"
# 	@echo "RabbitMQ setup complete! Admin user: admin/admin"

docker-build:
	sudo docker build -t lester -f ./lester/Dockerfile ./lester && \
	sudo docker build -t michael -f ./michael/Dockerfile ./michael && \
	sudo docker build -t franklin -f ./franklin/Dockerfile ./franklin && \
	sudo docker build -t trevor -f ./trevor/Dockerfile ./trevor

docker-build-lester:
	sudo docker build -t lester -f ./lester/Dockerfile ./lester

docker-build-michael:
	sudo docker build -t michael -f ./michael/Dockerfile ./michael

docker-build-franklin:
	sudo docker build -t franklin -f ./franklin/Dockerfile ./franklin

docker-build-trevor:
	sudo docker build -t trevor -f ./trevor/Dockerfile ./trevor

docker-run-lester:
	sudo docker run --name lester-container -p 50051:50051 lester

docker-run-michael:
	sudo docker run --name michael-container -p 50052:50052 michael

docker-run-franklin:
	sudo docker run --name franklin-container -p 50054:50054 franklin

docker-run-trevor:
	sudo docker run --name trevor-container -p 50053:50053 trevor

docker-logs-lester:
	sudo docker logs -f lester-container

docker-logs-michael:
	sudo docker logs -f michael-container

docker-logs-franklin:
	sudo docker logs -f franklin-container

docker-logs-trevor:
	sudo docker logs -f trevor-container

