package main

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestParseOptionsDefaults(t *testing.T) {
	withCommandLine(t, []string{"get-token"}, func() {
		opts := parseOptions()
		if opts == nil {
			t.Fatal("parseOptions() returned nil")
		}

		if opts.apiUrl != "https://api.vpsfree.cz" {
			t.Fatalf("apiUrl = %q", opts.apiUrl)
		}
		if opts.lifetime != "renewable_auto" {
			t.Fatalf("lifetime = %q", opts.lifetime)
		}
		if opts.interval != 3600 {
			t.Fatalf("interval = %d", opts.interval)
		}
		if opts.username != "" {
			t.Fatalf("username = %q", opts.username)
		}
		if opts.tfvars != "" {
			t.Fatalf("tfvars = %q", opts.tfvars)
		}
	})
}

func TestParseOptionsExplicitValues(t *testing.T) {
	withCommandLine(t, []string{
		"get-token",
		"-user", "alice",
		"-lifetime", "fixed",
		"-interval", "120",
		"-tfvars", "token.tfvars",
		"https://api.example.test",
	}, func() {
		opts := parseOptions()
		if opts == nil {
			t.Fatal("parseOptions() returned nil")
		}

		if opts.apiUrl != "https://api.example.test" {
			t.Fatalf("apiUrl = %q", opts.apiUrl)
		}
		if opts.username != "alice" {
			t.Fatalf("username = %q", opts.username)
		}
		if opts.lifetime != "fixed" {
			t.Fatalf("lifetime = %q", opts.lifetime)
		}
		if opts.interval != 120 {
			t.Fatalf("interval = %d", opts.interval)
		}
		if opts.tfvars != "token.tfvars" {
			t.Fatalf("tfvars = %q", opts.tfvars)
		}
	})
}

func TestParseOptionsRejectsTooManyArgs(t *testing.T) {
	withCommandLine(t, []string{
		"get-token",
		"https://api-one.example.test",
		"https://api-two.example.test",
	}, func() {
		if opts := parseOptions(); opts != nil {
			t.Fatalf("parseOptions() = %#v, want nil", opts)
		}
	})
}

func TestWriteTokenPrintsToken(t *testing.T) {
	out, err := captureStdout(t, func() error {
		return writeToken("secret-token", &options{})
	})
	if err != nil {
		t.Fatal(err)
	}
	if out != "secret-token" {
		t.Fatalf("stdout = %q, want token", out)
	}
}

func TestWriteTokenWritesTfvars(t *testing.T) {
	path := filepath.Join(t.TempDir(), "token.tfvars")

	if err := writeToken("secret-token", &options{tfvars: path}); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	want := `vpsadmin_token = "secret-token"`
	if string(b) != want {
		t.Fatalf("file = %q, want %q", string(b), want)
	}
}

func withCommandLine(t *testing.T, args []string, fn func()) {
	t.Helper()

	oldArgs := os.Args
	oldCommandLine := flag.CommandLine
	oldStderr := os.Stderr

	stderr, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatal(err)
	}

	os.Args = args
	os.Stderr = stderr
	flag.CommandLine = flag.NewFlagSet(args[0], flag.ContinueOnError)
	flag.CommandLine.SetOutput(stderr)

	defer func() {
		os.Args = oldArgs
		os.Stderr = oldStderr
		flag.CommandLine = oldCommandLine
		_ = stderr.Close()
	}()

	fn()
}

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = w
	fnErr := fn()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = oldStdout

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	return string(out), fnErr
}
