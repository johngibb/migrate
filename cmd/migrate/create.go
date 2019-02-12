package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
	"github.com/johngibb/migrate"
	"github.com/johngibb/migrate/source"
)

type Create struct {
	srcPath string
}

func (*Create) Name() string     { return "create" }
func (*Create) Synopsis() string { return "create new migration file" }
func (*Create) Usage() string {
	return `migrate create -src <folder> <migration name>:
    Creates a new migration file.
`
}

func (cmd *Create) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.srcPath, "src", ".", "directory containing migration files")
}

func (cmd *Create) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if len(f.Args()) < 1 {
		fmt.Fprint(os.Stderr, "error: missing migration name\n")
		f.Usage()
		return subcommands.ExitUsageError
	}
	src, err := source.New(cmd.srcPath)
	must(err)
	must(migrate.Create(src, f.Arg(0)))
	return subcommands.ExitSuccess
}
