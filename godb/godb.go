package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

// This is appropriate for `trust` authentication.
const url = "postgres://larstha@localhost:5432/example?sslmode=disable"

func main() {
	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	// Database "example"
	//   id     INTEGER PRIMARY
	//   name   VARCHAR(30)
	//   weight INTEGER

	scanner := bufio.NewScanner(os.Stdin)
Loop:
	for scanner.Scan() {
		xs := strings.Fields(scanner.Text())
		if len(xs) > 0 {
			switch xs[0] {
			case "h", "help":
				fmt.Println("a(ll)")
				fmt.Println("i(nsert) id name weight")
				fmt.Println("q(uery) id")
				fmt.Println("quit")
			case "quit":
				break Loop
			case "a", "all":
				rows, err := conn.Query(context.Background(), "select id, name, weight from widgets")
				if err != nil {
					fmt.Printf("Query failed: %v\n", err)
					continue Loop
				}
				for rows.Next() {
					var name string
					var id, weight int64
					rows.Scan(&id, &name, &weight)
					fmt.Println(id, name, weight)
				}
			case "i", "insert":
				if len(xs) != 4 {
					fmt.Println("Need three fields")
					continue Loop
				}
				// id name weight, the name can have no spaces
				id, err := strconv.ParseInt(xs[1], 10, 64)
				if err != nil {
					fmt.Println("bad id")
					continue Loop
				}
				name := xs[2]
				weight, err := strconv.ParseInt(xs[3], 10, 64)
				if err != nil || weight < 4 || weight > 300 {
					fmt.Println("bad weight")
					continue Loop
				}
				_, err = conn.Exec(context.Background(),
					"insert into widgets(id, name, weight) values ($1, $2, $3)",
					id, name, weight)
				if err != nil {
					fmt.Printf("Insert failed: %v\n", err)
					continue Loop
				}
				fmt.Println("Inserted")
			case "q", "query":
				if len(xs) != 2 {
					fmt.Println("Need one field")
					continue Loop
				}
				// id
				id, err := strconv.ParseInt(xs[1], 10, 64)
				if err != nil {
					fmt.Println("bad id")
					continue Loop
				}
				var name string
				var weight int64
				err = conn.QueryRow(context.Background(), "select name, weight from widgets where id=$1", id).Scan(&name, &weight)
				if err != nil {
					fmt.Printf("QueryRow failed: %v\n", err)
					continue Loop
				}
				fmt.Println(id, name, weight)
			default:
				fmt.Println("Try h for help")
			}
		}
	}
}
