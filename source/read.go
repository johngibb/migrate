package source

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"github.com/pkg/errors"
)

// ReadStatements reads the migration file and parses it into individual
// statements.
func (m *Migration) ReadStatements() ([]string, error) {
	f, err := m.fs.Open(m.Path)
	if err != nil {
		return nil, errors.Wrap(err, "could not open file")
	}
	defer f.Close()
	return splitStatements(f), nil
}

const (
	apostrophe = "'"
	dollarSign = "$"
	semicolon  = ";"
)

// splitStatements splits the given file into separate statements.
func splitStatements(r io.Reader) []string {
	var (
		buf      bytes.Buffer
		inBlock  bool
		inString bool
		result   []string
		scanner  = bufio.NewScanner(r)
	)

	scanner.Split(bufio.ScanBytes)
	last := ""

	for scanner.Scan() {
		curr := scanner.Text()
		_, _ = buf.WriteString(curr)

		switch last {
		case apostrophe:
			inString = !inString
		case dollarSign:
			if curr == dollarSign && !inString {
				inBlock = !inBlock
			}
		}

		if curr == semicolon && !(inBlock || inString) {
			result = append(result, buf.String())
			buf.Reset()
		}

		last = curr
	}

	if buf.Len() > 0 {
		if s := strings.TrimSpace(buf.String()); s != "" {
			result = append(result, s)
		}
	}

	return result
}
