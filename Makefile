run: build
	@./bin/webservice.exe

build:
	@go build -o bin/webservice.exe ./cmd/api/

dev:
	@air -c .air.toml

# seed information to database. Usefull for testing
seed:
	@go run cmd/seed/main.go

# go up to latest migration in database
up:
	@goose -dir ./sql/schema  sqlite3 ./app.db up

#go down a migration in database
down:
	@goose -dir ./sql/schema  sqlite3 ./app.db down 
# sqlc command. use when adding new sql queries
gen:
	@sqlc generate

test:
	@go test ./tests/
#linting with golangci-lint
lint:
	@golangci-lint run
#Check locally github workflows
ci/cd:
	@act
#Format and fix imports
fmt:
	@gofmt -s -w .
	@goimports -w .
