package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	pastepostgres "github.com/yueli-official/paste/internal/postgres"
)

func main() {
	databaseURL := os.Getenv("PASTE_DATABASE_URL")
	if databaseURL == "" {
		panic("PASTE_DATABASE_URL is required")
	}
	database, err := sql.Open("postgres", databaseURL)
	must(err)
	defer database.Close()
	ctx := context.Background()
	must(database.PingContext(ctx))
	must(pastepostgres.ApplySchema(ctx, database))
	fmt.Println("Paste schema is ready")
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
