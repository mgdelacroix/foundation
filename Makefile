check:
	go fmt ./...

test:
	go test -race -count 1 ./...
