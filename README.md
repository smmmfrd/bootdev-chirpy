# Bootdev Chirpy
Twitter clone project

## Commands
`go build -o out && ./out` to run the server

`goose postgres "DB_URL" up` Makes the migration for the database

`goose postgres "DB_URL" down` Undoes everything

`sqlc generate` makes the code in /internal for our databse queries.