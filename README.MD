install goose, sqlc, air

go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/air-verse/air@latest

todo reset jwt lookup token on logout, on pass change,
index the lookup_id column for fast lookups

golangci-lint to fix things

    "github.com/rs/zerolog/log"
