/*
---------------------------------------------------------------------------------------
File: repository.go
Autor: Aldenor
Data: 29-04-2026
----------------------------------------------------------------------------------------
Finalidade:
Especialização da interface "UserRepository" que generaliza as operações CRUD no banco de dados.
---------------------------------------------------------------------------------------
*/

package user

import (
	"context"
	"microsrv/internal/domain"
)

type UserRepository interface {
	domain.Repository[UserID, User, UserCreate, UserUpdate]
	SelectUserByName(ctx context.Context, username string) (*User, error)
	SelectByEmail(ctx context.Context, email string) (*User, error)
	SelectRows(ctx context.Context) ([]User, error)
	Search(ctx context.Context, filter UserSearch) ([]User, error)
}
