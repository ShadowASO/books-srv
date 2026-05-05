/*
---------------------------------------------------------------------------------------
File: books_repository.go
Autor: Aldenor
Data: 18-04-2026
----------------------------------------------------------------------------------------
Finalidade:
Especialização da interface Repository que generaliza as operações CRUD no banco de dados.
---------------------------------------------------------------------------------------
*/

package books

import (
	"context"
	"microsrv/internal/domain"
)

type BooksRepository interface {
	domain.Repository[BookID, Book, BookCreate, BookUpdate]
	SearchByNmObra(ctx context.Context, nmObra string, limit int64) ([]Book, error)
}
