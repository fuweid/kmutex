test: 
	go test -v ./.

proto:
	protoc -I=./examples/proto \
		./examples/proto/locker.proto \
		--go_out=plugins=grpc:./examples/proto

build-examples:
	go build -o bin/lockclient ./examples/lockclient
	go build -o bin/lockserver ./examples/lockserver
