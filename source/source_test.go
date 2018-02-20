package source

import (
	"reflect"
	"sort"
	"testing"
)

func TestParseMigration(t *testing.T) {
	tests := []struct {
		path string
		want *Migration
	}{{
		path: "./source/123_add_tables_to_db.sql",
		want: &Migration{
			Path:    "./source/123_add_tables_to_db.sql",
			Version: 123,
			Name:    "123_add_tables_to_db",
		},
	}}
	for i, tt := range tests {
		got, err := parseMigration(tt.path)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%d: got %v, want %v", i, got, tt.want)
		}
	}
}

func TestSort(t *testing.T) {
	tests := []struct {
		paths []string
		want  []string
	}{{
		paths: []string{
			"./source/10_tenth.sql",
			"./source/1_first.sql",
		},
		want: []string{
			"./source/1_first.sql",
			"./source/10_tenth.sql",
		},
	}}
	for i, tt := range tests {
		// Parse migrations.
		var migrations []*Migration
		for _, s := range tt.paths {
			m, err := parseMigration(s)
			if err != nil {
				t.Fatal("could not parse migration")
			}
			migrations = append(migrations, m)
		}

		// Sort migrations.
		sort.Sort(ByVersion(migrations))

		// Extract paths
		var got []string
		for _, m := range migrations {
			got = append(got, m.Path)
		}

		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%d: got %v, want %v", i, got, tt.want)
		}
	}
}
