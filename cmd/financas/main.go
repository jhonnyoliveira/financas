// Financas é o executável do sistema local de finanças pessoais.
package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/jhonnyoliveira/financas/internal/app"
	"github.com/jhonnyoliveira/financas/internal/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	code := cli.Execute(ctx, os.Args[1:], os.Stdout, os.Stderr, func(ctx context.Context, _ cli.Options) error {
		return app.Run(ctx)
	})
	os.Exit(code)
}
