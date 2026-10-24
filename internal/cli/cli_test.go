package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

func TestExecuteHelpShowsOptionsWithoutStartingApp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	called := false

	code := Execute(context.Background(), []string{"--help"}, &stdout, &stderr, func(context.Context, Options) error {
		called = true
		return nil
	})

	if code != 0 {
		t.Fatalf("Execute() code = %d, want 0", code)
	}
	if called {
		t.Fatal("Execute() iniciou a aplicação ao exibir ajuda")
	}
	for _, option := range []string{"--console", "--version", "--help"} {
		if !strings.Contains(stdout.String(), option) {
			t.Errorf("ajuda não contém %q: %q", option, stdout.String())
		}
	}
}

func TestExecuteVersionShowsVersionWithoutStartingApp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	called := false

	code := Execute(context.Background(), []string{"--version"}, &stdout, &stderr, func(context.Context, Options) error {
		called = true
		return nil
	})

	if code != 0 {
		t.Fatalf("Execute() code = %d, want 0", code)
	}
	if called {
		t.Fatal("Execute() iniciou a aplicação ao exibir versão")
	}
	if !strings.Contains(stdout.String(), "financas "+Version) {
		t.Errorf("versão inesperada: %q", stdout.String())
	}
}

func TestExecuteConsoleStartsApp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	var received Options

	code := Execute(context.Background(), []string{"--console"}, &stdout, &stderr, func(_ context.Context, options Options) error {
		received = options
		return nil
	})

	if code != 0 {
		t.Fatalf("Execute() code = %d, want 0", code)
	}
	if !received.Console {
		t.Fatal("Execute() não repassou a opção Console")
	}
}

func TestExecuteRejectsInvalidArgument(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := Execute(context.Background(), []string{"--invalida"}, &stdout, &stderr, func(context.Context, Options) error {
		t.Fatal("a aplicação não deve iniciar com opção inválida")
		return nil
	})

	if code != 2 {
		t.Fatalf("Execute() code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "flag provided but not defined") {
		t.Errorf("erro inesperado: %q", stderr.String())
	}
}

func TestExecuteReportsApplicationError(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := Execute(context.Background(), nil, &stdout, &stderr, func(context.Context, Options) error {
		return errors.New("falha de teste")
	})

	if code != 1 {
		t.Fatalf("Execute() code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "falha de teste") {
		t.Errorf("erro da aplicação não foi exibido: %q", stderr.String())
	}
}
