package buildtool

import (
	"testing"
)

func TestParseArguments(t *testing.T) {
	bt := BuildTarget{
		Build: Build{
			Inputs: []Input{"a", "b"},
		},
		BuildToolOptions: BuildToolOptions{
			Arguments: []BuildToolArgument{"{{ .Build.Inputs | Merge }}"},
		},
	}

	args := ParseArguments(bt, bt.Arguments...)
	ex := "a b"
	if len(args) != 2 {
		t.Fatalf("expected 2 arguments, got %d", len(args))
	}
	if slug := args.Slug(); slug != ex {
		t.Fatalf("expected '%s', got '%s'", ex, slug)
	}
}

func TestBuildToolArguments(t *testing.T) {
	bt := BuildTarget{
		Build: Build{
			Inputs: []Input{"a", "b"},
		},
		BuildToolOptions: BuildToolOptions{
			Arguments: []BuildToolArgument{"{{ .Build.Inputs | Merge }}"},
		},
	}
	btArgs := bt.BuildToolArguments()
	ex_0 := []string{"a", "b"}

	if len(btArgs) != len(ex_0) {
		t.Fatalf("expected %d args, got %d", len(ex_0), len(btArgs))
	}
	for i, ex_v := range ex_0 {
		if v := btArgs[i]; v != ex_v {
			t.Fatalf("exepected '%s', got '%s'", ex_v, v)
		}
	}
}

func TestBuildToolArgumentsWithStatics(t *testing.T) {
	bt := BuildTarget{
		Build: Build{
			Inputs: []Input{"a", "b"},
		},
		BuildToolOptions: BuildToolOptions{
			Arguments: []BuildToolArgument{"-o", "{{ .Build.Inputs | Merge }}"},
		},
	}
	btArgs := bt.BuildToolArguments()
	ex_0 := []string{"-o", "a", "b"}

	if len(btArgs) != len(ex_0) {
		t.Fatalf("expected %d args, got %d", len(ex_0), len(btArgs))
	}
	for i, ex_v := range ex_0 {
		if v := btArgs[i]; v != ex_v {
			t.Fatalf("exepected '%s', got '%s'", ex_v, v)
		}
	}
}
