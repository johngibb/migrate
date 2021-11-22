package main

import (
	"context"
	"flag"

	"github.com/google/subcommands"
	"github.com/johngibb/migrate"
	"github.com/johngibb/migrate/db"
	"github.com/johngibb/migrate/source"
)

type Up struct {
	conn    string
	srcPath string
	quiet   bool
}

func (*Up) Name() string     { return "up" }
func (*Up) Synopsis() string { return "apply all pending migrations to the db" }
func (*Up) Usage() string {
	return `migrate up -src <migrations folder> -conn <connection string> [-quiet]:
    Apply all pending migrations.
`
}

func (cmd *Up) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.conn, "conn", "", "postgres connection string")
	f.StringVar(&cmd.srcPath, "src", ".", "directory containing migration files")
	f.BoolVar(&cmd.quiet, "quiet", false, "only print errors")
}

func (cmd *Up) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	src, err := source.NewFromPath(cmd.srcPath)
	must(err)
	db, err := db.Connect(cmd.conn)
	must(err)
	defer db.Close()
	must(migrate.Up(src, db, cmd.quiet))
	return subcommands.ExitSuccess
}
