PROTOS = \
	agent.proto

gen: gen_go gen_py

gen_go:
	@for proto in $(PROTOS); do \
		echo $$proto; \
		protoc --proto_path=. --go_out=../data-orchestation/pb --go_opt=paths=source_relative \
		--go-grpc_out=../data-orchestation/pb --go-grpc_opt=paths=source_relative \
		$$proto; \
	done


gen_py:
	@for proto in $(PROTOS); do \
		echo $$proto; \
		python3 -m grpc_tools.protoc -I. --proto_path=. --python_out=../turtlebot-agent/pb --pyi_out=../turtlebot-agent/pb --grpc_python_out=../turtlebot-agent/pb $$proto; \
	done