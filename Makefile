lint:
	golangci-lint run

integration-test:
	go test -tags integration ./... -coverprofile=integration_coverage.out

e2e-test:
	go test -tags e2e ./... -coverprofile=e2e_coverage.out

run:
	go run ./cmd/api/server.go

run_db:
	docker-compose up postgres-db pgadmin-ui -d

restart_db:
	docker-compose down postgres-db pgadmin-ui
	make run_db