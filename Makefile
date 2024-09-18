lint:
	golangci-lint run

integration-test:
	go test -tags integration ./..