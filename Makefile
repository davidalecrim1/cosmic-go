lint:
	golangci-lint run

unit-test:
	go test -coverprofile=unit.out ./...

integration-test:
	$(MAKE) run-dependencies
	go test ./test/integration -tags integration -coverpkg=./internal/... -coverprofile=integration.out 

e2e-test:
	$(MAKE) run-dependencies
	go test -tags e2e ./... -coverpkg=./internal/... -coverprofile=e2e.out

coverage:
	gocovmerge unit.out integration.out e2e.out > combined.out
	go tool cover -html=coverage.out -o coverage.html

run:
	go run ./cmd/api/server.go

run-dependencies:
	docker-compose up postgres-db pgadmin-ui redis -d