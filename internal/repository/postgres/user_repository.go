/*
---------------------------------------------------------------------------------------
File: user_repository.go
Autor: Aldenor
Data: 29-04-2026
----------------------------------------------------------------------------------------
Finalidade:
Especialização da interface Repository que generaliza as operações CRUD no banco de dados.
---------------------------------------------------------------------------------------
*/
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"microsrv/internal/domain"
	"microsrv/internal/domain/user"
)

type UserPGRepository struct {
	Db *sql.DB
}

func NewUserPGRepository(db *sql.DB) *UserPGRepository {
	return &UserPGRepository{
		Db: db,
	}
}

func (r *UserPGRepository) ensureDB() error {
	if r == nil {
		return fmt.Errorf("repositório PostgreSQL não foi inicializado")
	}

	if r.Db == nil {
		return fmt.Errorf("conexão com o PostgreSQL não foi criada")
	}

	return nil
}

func normalizeUserRole(value string) string {
	role := strings.ToLower(strings.TrimSpace(value))

	switch role {
	case "admin", "developer":
		return role
	default:
		return "user"
	}
}

func (r *UserPGRepository) Insert(
	ctx context.Context,
	data user.UserCreate,
) (*domain.InsertResult[user.UserID], error) {
	if err := r.ensureDB(); err != nil {
		return nil, err
	}

	query := `
		INSERT INTO users (
			userrole,
			username,
			password,
			email,
			created_at
		)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING user_id
	`

	var userID user.UserID

	err := r.Db.QueryRowContext(
		ctx,
		query,
		normalizeUserRole(data.Userrole),
		strings.TrimSpace(data.Username),
		strings.TrimSpace(data.Password),
		strings.TrimSpace(data.Email),
	).Scan(&userID)

	if err != nil {
		return nil, fmt.Errorf("erro ao inserir usuário: %w", err)
	}

	return &domain.InsertResult[user.UserID]{
		ID: userID,
	}, nil
}

func (r *UserPGRepository) Update(
	ctx context.Context,
	userID user.UserID,
	data user.UserUpdate,
) (*domain.UpdateResult, error) {
	if err := r.ensureDB(); err != nil {
		return nil, err
	}

	query := `
		UPDATE users
		SET
			username = $1,
			userrole = $2,
			password = COALESCE(NULLIF($3, ''), password),
			email = $4
		WHERE user_id = $5
	`

	result, err := r.Db.ExecContext(
		ctx,
		query,
		strings.TrimSpace(data.Username),
		normalizeUserRole(data.Userrole),
		strings.TrimSpace(data.Password),
		strings.TrimSpace(data.Email),
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao atualizar usuário %v: %w", userID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter quantidade de linhas afetadas: %w", err)
	}

	return &domain.UpdateResult{
		MatchedCount:  rowsAffected,
		ModifiedCount: rowsAffected,
	}, nil
}

func (r *UserPGRepository) Delete(
	ctx context.Context,
	userID user.UserID,
) (*domain.DeleteResult, error) {
	if err := r.ensureDB(); err != nil {
		return nil, err
	}

	query := `
		DELETE FROM users
		WHERE user_id = $1
	`

	result, err := r.Db.ExecContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao excluir usuário %v: %w", userID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter quantidade de linhas excluídas: %w", err)
	}

	return &domain.DeleteResult{
		DeletedCount: rowsAffected,
	}, nil
}

func (r *UserPGRepository) Select(
	ctx context.Context,
	userID user.UserID,
) (*user.User, error) {
	if err := r.ensureDB(); err != nil {
		return nil, err
	}

	query := `
		SELECT
			user_id,
			userrole,
			username,
			password,
			email,
			created_at
			
		FROM users
		WHERE user_id = $1
	`

	var row user.User

	err := r.Db.QueryRowContext(ctx, query, userID).Scan(
		&row.UserID,
		&row.Userrole,
		&row.Username,
		&row.Password,
		&row.Email,
		&row.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("usuário não encontrado: %v", userID)
		}

		return nil, fmt.Errorf("erro ao consultar usuário %v: %w", userID, err)
	}

	return &row, nil
}

func (r *UserPGRepository) SelectUserByName(
	ctx context.Context,
	username string,
) (*user.User, error) {
	if err := r.ensureDB(); err != nil {
		return nil, err
	}

	query := `
		SELECT
			user_id,
			userrole,
			username,
			password,
			email,
			created_at
			
		FROM users
		WHERE username = $1
	`

	var row user.User

	err := r.Db.QueryRowContext(ctx, query, username).Scan(
		&row.UserID,
		&row.Userrole,
		&row.Username,
		&row.Password,
		&row.Email,
		&row.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("usuário não encontrado com username %q", username)
		}

		return nil, fmt.Errorf("erro ao consultar usuário por username %q: %w", username, err)
	}

	return &row, nil
}
func (r *UserPGRepository) SelectByEmail(
	ctx context.Context,
	email string,
) (*user.User, error) {
	if err := r.ensureDB(); err != nil {
		return nil, err
	}

	query := `
		SELECT
			user_id,
			userrole,
			username,
			password,
			email,
			created_at
			
		FROM users
		WHERE email = $1
	`

	var row user.User

	err := r.Db.QueryRowContext(ctx, query, email).Scan(
		&row.UserID,
		&row.Userrole,
		&row.Username,
		&row.Password,
		&row.Email,
		&row.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("usuário não encontrado com email %q", email)
		}

		return nil, fmt.Errorf("erro ao consultar usuário por email %q: %w", email, err)
	}

	return &row, nil
}
func (r *UserPGRepository) SelectRows(
	ctx context.Context,
) ([]user.User, error) {
	if err := r.ensureDB(); err != nil {
		return nil, err
	}

	query := `
		SELECT
			user_id,
			userrole,
			username,
			password,
			email,
			created_at		
		FROM users
		ORDER BY user_id ASC
	`

	rows, err := r.Db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("erro ao consultar usuários: %w", err)
	}
	defer rows.Close()

	var results []user.User

	for rows.Next() {
		var row user.User

		err := rows.Scan(
			&row.UserID,
			&row.Userrole,
			&row.Username,
			&row.Password,
			&row.Email,
			&row.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler usuário: %w", err)
		}

		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erro durante iteração dos usuários: %w", err)
	}

	return results, nil
}
func (r *UserPGRepository) Search(
	ctx context.Context,
	filter user.UserSearch,
) ([]user.User, error) {
	if err := r.ensureDB(); err != nil {
		return nil, err
	}

	limit := filter.NrDocs
	if limit <= 0 {
		limit = 50
	}

	search := strings.TrimSpace(filter.NmSearch)

	// Como a listagem geral já é feita por outro método,
	// busca vazia não deve retornar todos os registros.
	if search == "" {
		return []user.User{}, nil
	}

	query := `
		SELECT
			user_id,
			userrole,
			username,
			password,
			email,
			created_at
		FROM users
		WHERE
			username ILIKE '%' || $1 || '%'
		ORDER BY user_id ASC
		LIMIT $2
	`

	rows, err := r.Db.QueryContext(ctx, query, search, limit)
	if err != nil {
		return nil, fmt.Errorf("erro ao pesquisar usuários: %w", err)
	}
	defer rows.Close()

	results := make([]user.User, 0)

	for rows.Next() {
		var row user.User

		err := rows.Scan(
			&row.UserID,
			&row.Userrole,
			&row.Username,
			&row.Password,
			&row.Email,
			&row.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler usuário pesquisado: %w", err)
		}

		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erro durante iteração da pesquisa de usuários: %w", err)
	}

	return results, nil
}
