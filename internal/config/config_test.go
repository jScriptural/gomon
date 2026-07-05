package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()
	cmd := "./gomon"
	build := "go build -o app ."

	os.Args = []string{cmd, "--build", build}
	want := &Config{
		Build: build,
	}

	got, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if got.Build != want.Build {
		t.Fatalf("want: %#v but got: %#v", *want, *got)
	}
}
