package execx

import (
	"reflect"
	"testing"
)

func TestExpandArgsReplacesPlaceholders(t *testing.T) {
	template := []string{"cmd", "--model", "{model}", "--format", "{output_format}"}
	replacements := map[string]string{
		"{model}":         "sonnet",
		"{output_format}": "json",
	}
	got := ExpandArgs(template, replacements)
	want := []string{"cmd", "--model", "sonnet", "--format", "json"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ExpandArgs() = %v, want %v", got, want)
	}
}

func TestExpandArgsEmptyTemplate(t *testing.T) {
	got := ExpandArgs(nil, map[string]string{"{x}": "y"})
	if len(got) != 0 {
		t.Fatalf("expected empty result for nil template, got %v", got)
	}
}

func TestExpandArgsNoPlaceholders(t *testing.T) {
	template := []string{"git", "diff"}
	got := ExpandArgs(template, map[string]string{"{x}": "y"})
	if !reflect.DeepEqual(got, template) {
		t.Fatalf("ExpandArgs() = %v, want %v", got, template)
	}
}
