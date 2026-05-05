/*
---------------------------------------------------------------------------------------
File: crud_basico.go
Autor: Aldenor
Data: 18-04-2026
---------------------------------------------------------------------------------------
Finalidade:
Interface que generaliza as operações CRUD no banco de dados. A partir dela podemos
Criar tipos concretos ou especificar generalizações
---------------------------------------------------------------------------------------
*/
package domain

import "context"

type InsertResult[ID comparable] struct {
	ID ID
}

type UpdateResult struct {
	MatchedCount  int64
	ModifiedCount int64
}

type DeleteResult struct {
	DeletedCount int64
}

type Repository[ID comparable, TEntity any, TCreate any, TUpdate any] interface {
	Insert(ctx context.Context, entity TCreate) (*InsertResult[ID], error)
	Select(ctx context.Context, id ID) (*TEntity, error)
	Update(ctx context.Context, id ID, entity TUpdate) (*UpdateResult, error)
	Delete(ctx context.Context, id ID) (*DeleteResult, error)
}
