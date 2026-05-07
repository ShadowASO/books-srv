/*
---------------------------------------------------------------------------------------
File: user_handler.go
Autor: Aldenor
Data: 29-04-2026
---------------------------------------------------------------------------------------
Finalidade:
Handlers HTTP para operações CRUD de usuários.
---------------------------------------------------------------------------------------
*/

package handlers

import (
	"microsrv/internal/domain/user"

	"microsrv/internal/pkg/mslogger"
	"microsrv/internal/pkg/msresponse"
	"microsrv/internal/services"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	Service services.UserService
}

func NewUserHandler(srv services.UserService) *UserHandler {
	return &UserHandler{
		Service: srv,
	}
}

func validateUserIDParam(c *gin.Context) (user.UserID, bool) {
	idParam := strings.TrimSpace(c.Param("id"))

	if idParam == "" {
		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID não informado",
			msresponse.ErrorFormatoInvalido,
			"O parâmetro id é obrigatório.",
		)
		return 0, false
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil || id <= 0 {
		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID inválido",
			msresponse.ErrorFormatoInvalido,
			"O ID informado deve ser um número inteiro positivo.",
		)
		return 0, false
	}

	return user.UserID(id), true
}

// Insert godoc
//
// @Summary      Cadastra um novo usuário
// @Description  Insere um novo usuário na base PostgreSQL. Os campos userrole, username, password e email são obrigatórios.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      user.UserCreate  true  "Dados do usuário a ser cadastrado"
// @Success      201   {object}  msresponse.APIResponse "Usuário cadastrado com sucesso"
// @Failure      400   {object}  msresponse.APIResponse "JSON inválido ou campos obrigatórios ausentes"
// @Failure      500   {object}  msresponse.APIResponse "Erro interno ao cadastrar o usuário"
// @Router       /users [post]
func (h *UserHandler) Insert(c *gin.Context) {
	var body user.UserCreate

	if err := c.ShouldBindJSON(&body); err != nil {
		//mslogger.LoggerGlobal.Errorf("JSON com formato inválido: %v", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Formato inválido",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	body.Userrole = strings.TrimSpace(body.Userrole)
	body.Username = strings.TrimSpace(body.Username)
	body.Email = strings.TrimSpace(body.Email)

	if body.Userrole == "" || body.Username == "" || body.Password == "" || body.Email == "" {
		//mslogger.LoggerGlobal.Error("Faltam campos obrigatórios para cadastro de usuário")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Faltam campos obrigatórios",
			msresponse.ErrorValidacao,
			"Os campos userrole, username, password e email são obrigatórios.",
		)
		return
	}

	res, err := h.Service.Insert(c.Request.Context(), body)
	if err != nil {
		//mslogger.LoggerGlobal.Errorf("Erro ao inserir usuário: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro interno ao cadastrar o usuário",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	msresponse.OK(c, http.StatusCreated, "Usuário cadastrado com sucesso", res)
}

// SelectByID godoc
//
// @Summary      Busca um usuário por ID
// @Description  Retorna os dados de um usuário a partir do seu ID numérico.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID numérico do usuário"
// @Success      200  {object}  msresponse.APIResponse "Usuário localizado com sucesso"
// @Failure      400  {object}  msresponse.APIResponse "ID não informado ou inválido"
// @Failure      500  {object}  msresponse.APIResponse "Erro interno ao buscar o usuário"
// @Router       /users/{id} [get]
func (h *UserHandler) Select(c *gin.Context) {
	userID, ok := validateUserIDParam(c)
	if !ok {
		return
	}

	result, err := h.Service.Select(c.Request.Context(), userID)
	if err != nil {
		//mslogger.LoggerGlobal.Errorf("Erro ao buscar usuário id=%d: %v", userID, err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro interno ao buscar o usuário",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	msresponse.OK(c, http.StatusOK, "Usuário localizado com sucesso", result)
}

// SelectUserByName godoc
//
// @Summary      Busca um usuário por username
// @Description  Retorna os dados de um usuário a partir do username.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        username   path      string  true  "Username do usuário"
// @Success      200        {object}  msresponse.APIResponse "Usuário localizado com sucesso"
// @Failure      400        {object}  msresponse.APIResponse "Username não informado"
// @Failure      500        {object}  msresponse.APIResponse "Erro interno ao buscar o usuário"
// @Router       /users/username/{username} [get]
func (h *UserHandler) SelectUserByName(c *gin.Context) {
	username := strings.TrimSpace(c.Param("username"))

	if username == "" {
		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Username não informado",
			msresponse.ErrorFormatoInvalido,
			"O parâmetro username é obrigatório.",
		)
		return
	}

	result, err := h.Service.SelectUserByName(c.Request.Context(), username)
	if err != nil {
		//mslogger.LoggerGlobal.Errorf("Erro ao buscar usuário username=%s: %v", username, err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro interno ao buscar o usuário",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	msresponse.OK(c, http.StatusOK, "Usuário localizado com sucesso", result)
}

// SelectByEmail godoc
//
// @Summary      Busca um usuário por email
// @Description  Retorna os dados de um usuário a partir do email.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        email   path      string  true  "Email do usuário"
// @Success      200     {object}  msresponse.APIResponse "Usuário localizado com sucesso"
// @Failure      400     {object}  msresponse.APIResponse "Email não informado"
// @Failure      500     {object}  msresponse.APIResponse "Erro interno ao buscar o usuário"
// @Router       /users/email/{email} [get]
func (h *UserHandler) SelectByEmail(c *gin.Context) {
	email := strings.TrimSpace(c.Param("email"))

	if email == "" {
		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Email não informado",
			msresponse.ErrorFormatoInvalido,
			"O parâmetro email é obrigatório.",
		)
		return
	}

	result, err := h.Service.SelectByEmail(c.Request.Context(), email)
	if err != nil {
		//mslogger.LoggerGlobal.Errorf("Erro ao buscar usuário email=%s: %v", email, err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro interno ao buscar o usuário",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	msresponse.OK(c, http.StatusOK, "Usuário localizado com sucesso", result)
}

// SelectRows godoc
//
// @Summary      Lista usuários
// @Description  Retorna todos os usuários cadastrados.
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}  msresponse.APIResponse "Usuários listados com sucesso"
// @Failure      500  {object}  msresponse.APIResponse "Erro interno ao listar usuários"
// @Router       /users [get]
func (h *UserHandler) SelectRows(c *gin.Context) {
	result, err := h.Service.SelectRows(c.Request.Context())
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao listar usuários: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro interno ao listar usuários",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	msresponse.OK(c, http.StatusOK, "Usuários listados com sucesso", result)
}

// Update godoc
//
// @Summary      Atualiza um usuário por ID
// @Description  Atualiza os dados de um usuário existente. Se password vier vazio, a senha atual será mantida.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id    path      int              true  "ID numérico do usuário"
// @Param        user  body      user.UserUpdate  true  "Dados do usuário a serem atualizados"
// @Success      200   {object}  msresponse.APIResponse "Usuário atualizado com sucesso"
// @Failure      400   {object}  msresponse.APIResponse "ID inválido ou JSON em formato incorreto"
// @Failure      500   {object}  msresponse.APIResponse "Erro interno ao atualizar o usuário"
// @Router       /users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	userID, ok := validateUserIDParam(c)
	if !ok {
		return
	}

	var body user.UserUpdate

	if err := c.ShouldBindJSON(&body); err != nil {
		//mslogger.LoggerGlobal.Errorf("JSON com formato inválido: %v", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Formato inválido",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	body.Userrole = strings.TrimSpace(body.Userrole)
	body.Username = strings.TrimSpace(body.Username)
	body.Email = strings.TrimSpace(body.Email)

	if body.Userrole == "" || body.Username == "" || body.Email == "" {
		//mslogger.LoggerGlobal.Error("Faltam campos obrigatórios para atualização de usuário")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Faltam campos obrigatórios",
			msresponse.ErrorValidacao,
			"Os campos userrole, username e email são obrigatórios.",
		)
		return
	}

	result, err := h.Service.Update(c.Request.Context(), userID, body)
	if err != nil {
		//mslogger.LoggerGlobal.Errorf("Erro ao atualizar usuário id=%d: %v", userID, err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro interno ao atualizar o usuário",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	msresponse.OK(c, http.StatusOK, "Usuário atualizado com sucesso", result)
}

// Delete godoc
//
// @Summary      Remove um usuário por ID
// @Description  Exclui um usuário da base PostgreSQL a partir do seu ID numérico.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID numérico do usuário"
// @Success      200  {object}  msresponse.APIResponse "Usuário removido com sucesso"
// @Failure      400  {object}  msresponse.APIResponse "ID não informado ou inválido"
// @Failure      500  {object}  msresponse.APIResponse "Erro interno ao remover o usuário"
// @Router       /users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	userID, ok := validateUserIDParam(c)
	if !ok {
		return
	}

	result, err := h.Service.Delete(c.Request.Context(), userID)
	if err != nil {
		//mslogger.LoggerGlobal.Errorf("Erro ao remover usuário id=%d: %v", userID, err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro interno ao remover o usuário",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	msresponse.OK(c, http.StatusOK, "Usuário removido com sucesso", result)
}

// Search godoc
//
// @Summary      Pesquisa usuários
// @Description  Realiza busca aproximada de usuários por username, email ou userrole.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        search  body      user.UserSearch  true  "Parâmetros da busca"
// @Success      200     {object}  msresponse.APIResponse "Busca realizada com sucesso"
// @Failure      400     {object}  msresponse.APIResponse "JSON inválido"
// @Failure      500     {object}  msresponse.APIResponse "Erro interno ao buscar usuários"
// @Router       /users/search [post]
func (h *UserHandler) Search(c *gin.Context) {
	var filter user.UserSearch

	if err := c.ShouldBindJSON(&filter); err != nil {
		//mslogger.LoggerGlobal.Errorf("JSON com formato inválido: %v", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Formato inválido",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}
	//mslogger.LoggerGlobal.Errorf("JSON com formato inválido: %v", filter)

	filter.NmSearch = strings.TrimSpace(filter.NmSearch)

	if filter.NrDocs <= 0 {
		filter.NrDocs = 50
	}

	result, err := h.Service.Search(c.Request.Context(), filter)
	if err != nil {
		//mslogger.LoggerGlobal.Errorf("Erro ao pesquisar usuários: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro interno ao buscar usuários",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}
	//mslogger.LoggerGlobal.Infof("JSON com formato inválido: %v", result)

	msresponse.OK(c, http.StatusOK, "Busca realizada com sucesso", result)
}
