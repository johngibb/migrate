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
		},
		{
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
		{
			`create function update_trigger() returns trigger as $$
begin
  new.tsv :=
    to_tsvector(coalesce(new.alpha, 'foo''s')) ||
    to_tsvector(coalesce(new.bravo, '$$'));
  return new;
end
$$ language plpgsql;`,
			nil,
		},
		{
			`select * from table where thing not in (';''', '');
select * from table where thing not in ('', '''');`,
			[]string{
				`select * from table where thing not in (';''', '');`,
				`select * from table where thing not in ('', '''');`,
			},
		},
	}
	for i, tt := range tests {
		want := tt.statements
		if len(want) == 0 {
			want = []string{tt.src}
		}
		if got := trimAll(splitStatements(strings.NewReader(tt.src))); !reflect.DeepEqual(got, want) {
			t.Errorf("read failed: %d, got: \n%s\nwant: \n%s", i, toJSON(got), toJSON(tt.statements))
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

func toJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}
