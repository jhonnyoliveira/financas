// Package app contém o ciclo de vida da aplicação.
package app

import "context"

// Run mantém a aplicação em execução até que o contexto seja cancelado.
// Nesta etapa não há serviços de negócio a inicializar.
func Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
