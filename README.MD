install goose, sqlc, air

go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/air-verse/air@latest

todo reset jwt lookup token on logout, on pass change,
index the lookup_id column for fast lookups

golangci-lint to fix things

get user from context do be fked. FIX ASAP

// Consider caching user data to reduce database hits
var userCache = make(map[string]db.User)

Calling a function to log inside logger will break caller info

Log sampling with Zerolog
Sampling is a technique used to intentionally drop repetitive log entries so that only a proportion of them are kept and processed without sacrificing the insights derived from the logs. This is helpful when your highly trafficked application is producing a massive amount of records, and storing every single one will lead to excessive storage and processing costs which may not be desirable. Sampling fixes this problem by preventing the same logs from being recorded hundreds or thousands of times per second which prevents resources from being used up unnecessarily.

Here's the most basic way to sample logs with Zerolog:

 

func main() {
    log := zerolog.New(os.Stdout).
        With().
        Timestamp().
        Logger().
        Sample(&zerolog.BasicSampler{N: 5})

    for i := 1; i <= 10; i++ {
        log.Info().Msgf("a message from the gods: %d", i)
    }
}
