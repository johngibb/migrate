package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/google/subcommands"
)

func main() {
	log.SetFlags(0)
	subcommands.Register(&Status{}, "")
	subcommands.Register(&Up{}, "")
	subcommands.Register(&Create{}, "")
	subcommands.Register(subcommands.HelpCommand(), "")

	os.Args = translateLegacyArgs(os.Args)

	flag.Parse()
	os.Exit(int(
		subcommands.Execute(context.Background()),
	))
}

// must calls log.Fatal if the error is non-nil.
func must(err error) {
	if err != nil {
		log.Fatalf("migrate: %v\n", err)
	}
}

// translateLegacyArgs translates legacy command-line arguments to the newer
// format used by subcommands. Formerly, flags were specified before the
// command.
//
// Example:
// 	 before: [migrate -src ./migrations create add_table]
// 	 after:  [migrate create -src ./migrations add_table]
func translateLegacyArgs(args []string) []string {
	if len(args) < 2 {
		return args
	}
	var (
		cmd   string
		flags []string
	)
	if !isFlag(args[1]) {
		return args
	}
	for _, arg := range args[1:] {
		if isCommand(arg) {
			cmd = arg
		} else {
			flags = append(flags, arg)
		}
	}

	// Special case: "create" no longer accepts a "conn" flag.
	if cmd == "create" {
		for i, s := range flags {
			switch {
			case strings.HasPrefix(s, "-conn="), strings.HasPrefix(s, "--conn="):
				flags = append(flags[0:i], flags[i+1:]...)
			case strings.HasPrefix(s, "-conn"), strings.HasPrefix(s, "--conn"):
				flags = append(flags[0:i], flags[i+2:]...)
				break
			}
		}
	}
	return append([]string{args[0], cmd}, flags...)
}

func isFlag(s string) bool {
	return strings.HasPrefix(s, "-conn") || strings.HasPrefix(s, "--conn") ||
		strings.HasPrefix(s, "-src") || strings.HasPrefix(s, "--src") ||
		strings.HasPrefix(s, "-quiet") || strings.HasPrefix(s, "--quiet")
}

func isCommand(s string) bool {
	return s == "create" ||
		s == "status" ||
		s == "up"
}
