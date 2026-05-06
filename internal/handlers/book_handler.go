/*
---------------------------------------------------------------------------------------
File: book_Handler.go
Autor: Aldenor
Data: 15-04-2026
---------------------------------------------------------------------------------------
*/
package handlers

import (
	"microsrv/internal/domain/books"

	"microsrv/internal/pkg/msresponse"
	"microsrv/internal/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type BooksHandler struct {
	Service services.BooksService
}

func NewBooksHandler(srv services.BooksService) *BooksHandler {
	return &BooksHandler{Service: srv}
}

func validateObjectIDParam(c *gin.Context) (string, bool) {
	idHex := strings.TrimSpace(c.Param("id"))

	if idHex == "" {
		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID não informado",
			msresponse.ErrorFormatoInvalido,
			"O parâmetro id é obrigatório.",
		)
		return "", false
	}

	if _, err := bson.ObjectIDFromHex(idHex); err != nil {
		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID inválido",
			msresponse.ErrorFormatoInvalido,
			"O ID informado não é um ObjectID válido.",
		)
		return "", false
	}

	return idHex, true
}

// Insert godoc
//
// @Summary      Cadastra um novo livro
// @Description  Insere um novo registro de livro na base de dados. Os campos nome da obra e autor são obrigatórios.
// @Tags         books
// @Accept       json
// @Produce      json
// @Param        book  body      books.BookCreate  true  "Dados do livro a ser cadastrado"
// @Success      201   {object}  msresponse.APIResponse "Livro cadastrado com sucesso"
// @Failure      400   {object}  msresponse.APIResponse "JSON inválido ou campos obrigatórios ausentes"
// @Failure      500   {object}  msresponse.APIResponse "Erro interno ao cadastrar o livro"
// @Router       /tabelas/books [post]
func (h *BooksHandler) Insert(c *gin.Context) {
	var book books.BookCreate

	if err := c.ShouldBindJSON(&book); err != nil {
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

	if strings.TrimSpace(book.NmObra) == "" || strings.TrimSpace(book.NmAutor) == "" {
		//mslogger.LoggerGlobal.Error("Faltam campos obrigatórios")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Faltam campos obrigatórios",
			msresponse.ErrorValidacao,
			"Os campos nm_obra e nm_autor são obrigatórios.",
		)
		return
	}

	res, err := h.Service.Insert(c.Request.Context(), book)
	if err != nil {
		//mslogger.LoggerGlobal.Errorf("Erro ao inserir livro: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro interno ao cadastrar o livro",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	msresponse.OK(c, http.StatusCreated, "Livro cadastrado com sucesso", res)
}

// Select godoc
//
// @Summary      Busca um livro por ID
// @Description  Retorna os dados de um livro a partir do seu identificador MongoDB ObjectID.
// @Tags         books
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "ID do livro no formato ObjectID hexadecimal"
// @Success      200  {object}  msresponse.APIResponse "Livro localizado com sucesso"
// @Failure      400  {object}  msresponse.APIResponse "ID não informado ou inválido"
// @Failure      500  {object}  msresponse.APIResponse "Erro interno ao buscar o livro"
// @Router       /tabelas/books/{id} [get]
func (h *BooksHandler) Select(c *gin.Context) {
	idHex, ok := validateObjectIDParam(c)
	if !ok {
		return
	}

	result, err := h.Service.Select(c.Request.Context(), books.BookID(idHex))
	if err != nil {
		//mslogger.LoggerGlobal.Errorf("Erro ao buscar livro id=%s: %v", idHex, err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro interno ao buscar o livro",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	msresponse.OK(c, http.StatusOK, "Livro localizado com sucesso", result)
}

// Delete godoc
//
// @Summary      Remove um livro por ID
// @Description  Exclui um livro da base de dados a partir do seu identificador MongoDB ObjectID.
// @Tags         books
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "ID do livro no formato ObjectID hexadecimal"
// @Success      200  {object}  msresponse.APIResponse "Livro removido com sucesso"
// @Failure      400  {object}  msresponse.APIResponse "ID não informado ou inválido"
// @Failure      500  {object}  msresponse.APIResponse "Erro interno ao remover o livro"
// @Router       /tabelas/books/{id} [delete]
func (h *BooksHandler) Delete(c *gin.Context) {
	idHex, ok := validateObjectIDParam(c)
	if !ok {
		return
	}

	result, err := h.Service.Delete(c.Request.Context(), books.BookID(idHex))
	if err != nil {
		//mslogger.LoggerGlobal.Errorf("Erro ao remover livro id=%s: %v", idHex, err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro interno ao remover o livro",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	msresponse.OK(c, http.StatusOK, "Livro removido com sucesso", result)
}

// Update godoc
//
// @Summary      Atualiza um livro por ID
// @Description  Atualiza os dados de um livro existente a partir do seu identificador MongoDB ObjectID.
// @Tags         books
// @Accept       json
// @Produce      json
// @Param        id    path      string            true  "ID do livro no formato ObjectID hexadecimal"
// @Param        book  body      books.BookUpdate  true  "Dados do livro a serem atualizados"
// @Success      200   {object}  msresponse.APIResponse "Livro atualizado com sucesso"
// @Failure      400   {object}  msresponse.APIResponse "ID inválido ou JSON em formato incorreto"
// @Failure      500   {object}  msresponse.APIResponse "Erro interno ao atualizar o livro"
// @Router       /tabelas/books/{id} [put]
func (h *BooksHandler) Update(c *gin.Context) {
	idHex, ok := validateObjectIDParam(c)
	if !ok {
		return
	}

	var book books.BookUpdate

	if err := c.ShouldBindJSON(&book); err != nil {
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

	result, err := h.Service.Update(c.Request.Context(), books.BookID(idHex), book)
	if err != nil {
		//mslogger.LoggerGlobal.Errorf("Erro ao atualizar livro id=%s: %v", idHex, err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro interno ao atualizar o livro",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	msresponse.OK(c, http.StatusOK, "Livro atualizado com sucesso", result)
}

// SearchByNmObra godoc
//
// @Summary      Busca livros pelo nome da obra
// @Description  Realiza busca aproximada de livros pelo nome da obra. Caso o número máximo de documentos não seja informado, será utilizado o limite padrão de 10 registros.
// @Tags         books
// @Accept       json
// @Produce      json
// @Param        search  body      books.BookSearch  true  "Parâmetros da busca aproximada"
// @Success      200     {object}  msresponse.APIResponse "Busca realizada com sucesso"
// @Failure      400     {object}  msresponse.APIResponse "JSON inválido ou expressão de busca não informada"
// @Failure      500     {object}  msresponse.APIResponse "Erro interno ao buscar livros"
// @Router       /tabelas/books/search/nm-obra [post]
func (h *BooksHandler) SearchByNmObra(c *gin.Context) {
	var book books.BookSearch

	if err := c.ShouldBindJSON(&book); err != nil {
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

	book.NmSearch = strings.TrimSpace(book.NmSearch)

	if book.NmSearch == "" {
		//mslogger.LoggerGlobal.Error("Expressão de busca não informada")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Expressão de busca não informada",
			msresponse.ErrorValidacao,
			"O campo nm_search é obrigatório.",
		)
		return
	}

	if book.NrDocs <= 0 {
		book.NrDocs = 10
	}

	result, err := h.Service.SearchByNmObra(
		c.Request.Context(),
		book.NmSearch,
		int64(book.NrDocs),
	)
	if err != nil {
		//mslogger.LoggerGlobal.Errorf("Erro ao buscar livros por nome da obra: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro interno ao buscar livros",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	msresponse.OK(c, http.StatusOK, "Busca realizada com sucesso", result)
}
