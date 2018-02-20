package source

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
)

// ReadStatements reads the migration file and parses it into individual
// statements.
func (m *Migration) ReadStatements() ([]string, error) {
	f, err := os.Open(m.Path)
	if err != nil {
		return nil, errors.Wrap(err, "could not open file")
	}
	defer f.Close()
	return splitStatements(f), nil
}

// splitStatements splits the given file into separate statements.
func splitStatements(r io.Reader) []string {
	var buf bytes.Buffer
	scanner := bufio.NewScanner(r)
	var result []string
	for scanner.Scan() {
		l := scanner.Text()
		_, _ = buf.WriteString(l + "\n")
		if strings.HasSuffix(l, ";") {
			result = append(result, buf.String())
			buf.Reset()
		}
	}
	if buf.Len() > 0 {
		if s := strings.TrimSpace(buf.String()); s != "" {
			result = append(result, buf.String())
		}
	}
	return result
}
