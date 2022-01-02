install:
	go install ./cmd/pbr

example:
	cd ./internal/example/helloworld && go run main.go
