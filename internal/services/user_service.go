/*
---------------------------------------------------------------------------------------
File: user_service.go
Autor: Aldenor
Data: 29-04-2026
----------------------------------------------------------------------------------------
Finalidade:
Operações de serviço para usuários, utilizando repositório PostgreSQL.
---------------------------------------------------------------------------------------
*/

package services

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"microsrv/internal/domain"
	"microsrv/internal/domain/user"
	"microsrv/internal/pkg/mslogger"
)

type UserService struct {
	Repo user.UserRepository
}

var UserServiceGlobal *UserService
var onceInitUserService sync.Once

func InitUserService(repo user.UserRepository) {
	onceInitUserService.Do(func() {
		UserServiceGlobal = &UserService{
			Repo: repo,
		}

		mslogger.LoggerGlobal.Info("Global UserService configurado com sucesso.")
	})
}

func NewUserService(repo user.UserRepository) *UserService {
	return &UserService{
		Repo: repo,
	}
}

func (s *UserService) GetUserRepository() (user.UserRepository, error) {
	if s == nil {
		mslogger.LoggerGlobal.Error("UserService não foi inicializado corretamente: service é nil.")
		return nil, fmt.Errorf("serviço de usuário não inicializado")
	}

	if s.Repo == nil {
		mslogger.LoggerGlobal.Error("UserService não foi inicializado corretamente: repositório é nil.")
		return nil, fmt.Errorf("repositório de usuário não inicializado")
	}

	return s.Repo, nil
}

func (s *UserService) Insert(
	ctx context.Context,
	body user.UserCreate,
) (*domain.InsertResult[user.UserID], error) {
	repo, err := s.GetUserRepository()
	if err != nil {
		return nil, err
	}

	data := user.UserCreate{
		Userrole: strings.TrimSpace(body.Userrole),
		Username: strings.TrimSpace(body.Username),
		Password: body.Password,
		Email:    strings.TrimSpace(body.Email),
	}

	if data.Userrole == "" {
		return nil, fmt.Errorf("perfil do usuário não informado")
	}

	if data.Username == "" {
		return nil, fmt.Errorf("nome de usuário não informado")
	}

	if data.Password == "" {
		return nil, fmt.Errorf("senha não informada")
	}

	if data.Email == "" {
		return nil, fmt.Errorf("email não informado")
	}

	userID, err := repo.Insert(ctx, data)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao inserir usuário %q: %v", data.Username, err)
		return nil, fmt.Errorf("erro ao inserir usuário: %w", err)
	}

	mslogger.LoggerGlobal.Infof("usuário inserido com sucesso: %d", userID)

	return &domain.InsertResult[user.UserID]{
		ID: userID.ID,
	}, nil
}

func (s *UserService) Select(
	ctx context.Context,
	userID user.UserID,
) (*user.User, error) {
	repo, err := s.GetUserRepository()
	if err != nil {
		return nil, err
	}

	if userID <= 0 {
		return nil, fmt.Errorf("ID de usuário inválido: %d", userID)
	}

	result, err := repo.Select(ctx, userID)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao selecionar usuário ID %d: %v", userID, err)
		return nil, err
	}

	mslogger.LoggerGlobal.Infof("usuário selecionado com sucesso: %d", userID)

	return result, nil
}

func (s *UserService) SelectUserByName(
	ctx context.Context,
	username string,
) (*user.User, error) {
	repo, err := s.GetUserRepository()
	if err != nil {
		return nil, err
	}

	username = strings.TrimSpace(username)
	if username == "" {
		return nil, fmt.Errorf("nome de usuário não informado")
	}

	result, err := repo.SelectUserByName(ctx, username)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao selecionar usuário por username %q: %v", username, err)
		return nil, err
	}

	mslogger.LoggerGlobal.Infof("usuário selecionado com sucesso por username: %s", username)

	return result, nil
}

func (s *UserService) SelectByEmail(
	ctx context.Context,
	email string,
) (*user.User, error) {
	repo, err := s.GetUserRepository()
	if err != nil {
		return nil, err
	}

	email = strings.TrimSpace(email)
	if email == "" {
		return nil, fmt.Errorf("email não informado")
	}

	result, err := repo.SelectByEmail(ctx, email)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao selecionar usuário por email %q: %v", email, err)
		return nil, err
	}

	mslogger.LoggerGlobal.Infof("usuário selecionado com sucesso por email: %s", email)

	return result, nil
}

func (s *UserService) SelectRows(
	ctx context.Context,
) ([]user.User, error) {
	repo, err := s.GetUserRepository()
	if err != nil {
		return nil, err
	}

	results, err := repo.SelectRows(ctx)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao consultar usuários: %v", err)
		return nil, fmt.Errorf("erro ao consultar usuários: %w", err)
	}

	mslogger.LoggerGlobal.Infof("consulta de usuários realizada com sucesso. Total: %d", len(results))

	return results, nil
}

func (s *UserService) Search(
	ctx context.Context,
	filter user.UserSearch,
) ([]user.User, error) {
	repo, err := s.GetUserRepository()
	if err != nil {
		return nil, err
	}

	filter.NmSearch = strings.TrimSpace(filter.NmSearch)

	if filter.NrDocs <= 0 {
		filter.NrDocs = 50
	}

	results, err := repo.Search(ctx, filter)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao pesquisar usuários por %q: %v", filter.NmSearch, err)
		return nil, fmt.Errorf("erro ao pesquisar usuários: %w", err)
	}

	mslogger.LoggerGlobal.Infof(
		"pesquisa de usuários realizada com sucesso. termo=%q total=%d",
		filter.NmSearch,
		len(results),
	)

	return results, nil
}

func (s *UserService) Update(
	ctx context.Context,
	userID user.UserID,
	body user.UserUpdate,
) (*domain.UpdateResult, error) {
	repo, err := s.GetUserRepository()
	if err != nil {
		return nil, err
	}

	if userID <= 0 {
		return nil, fmt.Errorf("ID de usuário inválido: %d", userID)
	}

	data := user.UserUpdate{
		Userrole: strings.TrimSpace(body.Userrole),
		Username: strings.TrimSpace(body.Username),
		Password: body.Password,
		Email:    strings.TrimSpace(body.Email),
	}

	if data.Userrole == "" {
		return nil, fmt.Errorf("perfil do usuário não informado")
	}

	if data.Username == "" {
		return nil, fmt.Errorf("nome de usuário não informado")
	}

	if data.Email == "" {
		return nil, fmt.Errorf("email não informado")
	}

	resp, err := repo.Update(ctx, userID, data)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao alterar usuário ID %d: %v", userID, err)
		return nil, fmt.Errorf("erro ao alterar usuário ID %d: %w", userID, err)
	}

	if resp.ModifiedCount == 0 {
		mslogger.LoggerGlobal.Errorf("nenhum usuário encontrado para alteração com ID %d", userID)
		return nil, fmt.Errorf("nenhum usuário encontrado para alteração com ID %d", userID)
	}

	mslogger.LoggerGlobal.Infof("usuário alterado com sucesso: %d", userID)

	return &domain.UpdateResult{
		MatchedCount:  resp.MatchedCount,
		ModifiedCount: resp.ModifiedCount,
	}, nil
}

func (s *UserService) Delete(
	ctx context.Context,
	userID user.UserID,
) (*domain.DeleteResult, error) {
	repo, err := s.GetUserRepository()
	if err != nil {
		return nil, err
	}

	if userID <= 0 {
		return nil, fmt.Errorf("ID de usuário inválido: %d", userID)
	}

	resp, err := repo.Delete(ctx, userID)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao deletar usuário ID %d: %v", userID, err)
		return nil, fmt.Errorf("erro ao deletar usuário ID %d: %w", userID, err)
	}

	if resp.DeletedCount == 0 {
		mslogger.LoggerGlobal.Errorf("nenhum usuário encontrado para exclusão com ID %d", userID)
		return nil, fmt.Errorf("nenhum usuário encontrado para exclusão com ID %d", userID)
	}

	mslogger.LoggerGlobal.Infof("usuário deletado com sucesso: %d", userID)

	return &domain.DeleteResult{
		DeletedCount: resp.DeletedCount,
	}, nil
}
