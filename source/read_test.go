package source

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestSplitStatements(t *testing.T) {
	tests := []struct {
		src        string
		statements []string
	}{
		{
			`create table test(id int)`,
			[]string{`create table test(id int)`},
		},
		{
			`
                create table test1(id int);
                create table test2(id int);
            `,
			[]string{
				`create table test1(id int);`,
				`create table test2(id int);`,
			},
		}, {
			`
                create table test1(
                    id int,
                    name text
                );
                create table test2(id int);
            `,
			[]string{
				`create table test1(
                    id int,
                    name text
                );`,
				`create table test2(id int);`,
			},
		},
		{
			`insert into test select ';'`,
			[]string{`insert into test select ';'`},
		},
	}
	for i, tt := range tests {
		if got := trimAll(splitStatements(strings.NewReader(tt.src))); !reflect.DeepEqual(got, tt.statements) {
			t.Errorf("read failed: %d, got: \n%s\nwant: \n%s", i, toJson(got), toJson(tt.statements))
		}
	}

}

func trimAll(ss []string) []string {
	result := make([]string, len(ss))
	for i, s := range ss {
		result[i] = strings.TrimSpace(s)
	}
	return result
}

func toJson(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}
