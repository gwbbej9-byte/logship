package main

import (
	"reflect"
	"testing"
)

func TestSplitTailArgs(t *testing.T) {
	files, flags, err := splitTailArgs([]string{
		"samples/app.log", "samples/api.log", "--listen", ":8080", "--json-out",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(files, []string{"samples/app.log", "samples/api.log"}) {
		t.Fatalf("files=%v", files)
	}
	if !reflect.DeepEqual(flags, []string{"--listen", ":8080", "--json-out"}) {
		t.Fatalf("flags=%v", flags)
	}
}

func TestSplitTailArgsLevel(t *testing.T) {
	files, flags, err := splitTailArgs([]string{"a.log", "--level", "error,warn"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(files, []string{"a.log"}) {
		t.Fatalf("files=%v", files)
	}
	if !reflect.DeepEqual(flags, []string{"--level", "error,warn"}) {
		t.Fatalf("flags=%v", flags)
	}
}

func TestSplitTailArgsListenEquals(t *testing.T) {
	files, flags, err := splitTailArgs([]string{"--listen=:8080", "a.log"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(files, []string{"a.log"}) {
		t.Fatalf("files=%v", files)
	}
	if !reflect.DeepEqual(flags, []string{"--listen=:8080"}) {
		t.Fatalf("flags=%v", flags)
	}
}
