package main

import (
	"reflect"
	"testing"
)

func TestTranslateLegacyArgs(t *testing.T) {
	tests := []struct {
		args []string
		want []string
	}{
		{
			[]string{"migrate", "--src", "./dir", "up"},
			[]string{"migrate", "up", "--src", "./dir"},
		},
		{
			[]string{"migrate", "-src", "./migs", "-quiet", "up"},
			[]string{"migrate", "up", "-src", "./migs", "-quiet"},
		},
		{
			[]string{"migrate", "-src", "./migs", "-conn", "myconn", "create"},
			[]string{"migrate", "create", "-src", "./migs"},
		},
	}

	for _, tt := range tests {
		got := translateLegacyArgs(tt.args)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("got %v, want %v", got, tt.want)
		}
	}
}
