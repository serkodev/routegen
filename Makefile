install:
	go install ./cmd/routegen

example:
	cd ./internal/example/helloworld && go run main.go
