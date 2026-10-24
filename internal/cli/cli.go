// Package cli implementa a interface de linha de comando do executável.
package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
)

// Version identifica a versão do executável. Pode ser substituída no build.
var Version = "dev"

// Options reúne as opções suportadas pelo executável.
type Options struct {
	Console bool
}

// Runner inicia a aplicação com as opções informadas.
type Runner func(context.Context, Options) error

// Execute interpreta os argumentos e executa a aplicação quando apropriado.
// Retorna um código de saída compatível com a linha de comando.
func Execute(ctx context.Context, args []string, stdout, stderr io.Writer, run Runner) int {
	flags := flag.NewFlagSet("financas", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() { printUsage(stdout) }

	console := flags.Bool("console", false, "executa em modo console")
	version := flags.Bool("version", false, "exibe a versão e encerra")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(stderr, "erro: %v\n", err)
		printUsage(stderr)
		return 2
	}

	if *version {
		fmt.Fprintf(stdout, "financas %s\n", Version)
		return 0
	}

	if flags.NArg() != 0 {
		fmt.Fprintf(stderr, "erro: argumentos não reconhecidos: %v\n", flags.Args())
		printUsage(stderr)
		return 2
	}

	if err := run(ctx, Options{Console: *console}); err != nil {
		fmt.Fprintf(stderr, "erro ao executar a aplicação: %v\n", err)
		return 1
	}
	return 0
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "Uso: financas [opções]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Opções:")
	fmt.Fprintln(w, "  --console  executa em modo console")
	fmt.Fprintln(w, "  --version  exibe a versão e encerra")
	fmt.Fprintln(w, "  --help     exibe esta ajuda")
}
