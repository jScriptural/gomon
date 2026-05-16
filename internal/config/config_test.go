package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()
	cmd := "./gomon"
	preBuild := "go vet ./..."
	build := "go build -o app ."

	os.Args = []string{cmd, "--prebuild", preBuild, "--build", build}
	want := &Config{
		Build: build,
		Hooks: Hooks{
			PreBuild: preBuild,
		},
	}
	got, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if got.Build != want.Build || got.Hooks.PreBuild != want.Hooks.PreBuild {
		t.Fatalf("want: %#v but got: %#v", *want, *got)
	}
}
