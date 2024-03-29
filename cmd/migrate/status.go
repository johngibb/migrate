package main

import (
	"context"
	"flag"

	"github.com/google/subcommands"

	"github.com/johngibb/migrate"
	"github.com/johngibb/migrate/db"
	"github.com/johngibb/migrate/source"
)

type Status struct {
	conn    string
	srcPath string
}

func (*Status) Name() string     { return "status" }
func (*Status) Synopsis() string { return "display the current status of the migrations" }
func (*Status) Usage() string {
	return `migrate status:
    Display a list of pending and applied migrations.
`
}

func (cmd *Status) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.conn, "conn", "", "postgres connection string")
	f.StringVar(&cmd.srcPath, "src", ".", "directory containing migration files")
}

func (cmd *Status) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	src, err := source.New(cmd.srcPath)
	must(err)
	db, err := db.Connect(ctx, cmd.conn)
	must(err)
	defer db.Close(ctx)
	must(migrate.Status(ctx, src, db))
	return subcommands.ExitSuccess
}
