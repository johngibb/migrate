package main

import (
	"flag"
	"log"

	"github.com/johngibb/migrate"
	"github.com/johngibb/migrate/db"
	"github.com/johngibb/migrate/source"
)

func main() {
	var (
		conn    = flag.String("conn", "", "postgres connection string")
		srcPath = flag.String("src", ".", "directory containing migration files")
	)
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		flag.Usage()
		return
	}

	src, err := source.New(*srcPath)
	must(err)
	db, err := db.Connect(*conn)
	must(err)
	defer db.Close()

	// Dispatch command.
	switch args[0] {
	case "-h", "--help":
		flag.Usage()
		return
	case "create":
		if len(args) < 2 {
			flag.Usage()
			return
		}
		must(migrate.Create(src, args[1]))
	case "status":
		must(migrate.Status(src, db))
	case "up":
		must(migrate.Up(src, db))
	default:
		flag.Usage()
		return
	}
}

// must calls log.Fatal if the error is non-nil.
func must(err error) {
	if err != nil {
		log.Fatalf("migrate: %v\n", err)
	}
}

// usage prints usage information.
func usage() {
	log.Print(usagePrefix)
	flag.PrintDefaults()
	log.Print(usageCommands)
}

var (
	usagePrefix = `Usage: migrate [options] command

Options:
`

	usageCommands = `
Commands:
    create NAME    create new migration file
    status         display the current status of the migrations
    up             apply all pending migrations to the db
`
)
