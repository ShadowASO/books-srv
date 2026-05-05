/*
---------------------------------------------------------------------------------------
File: book_service.go
Autor: Aldenor
Data: 14-04-2026
----------------------------------------------------------------------------------------
Finalidade:
Operações CRUD na Collection "books" do MongoDB
---------------------------------------------------------------------------------------
*/
package services

import (
	"context"
	"errors"
	"fmt"

	"microsrv/internal/domain"
	"microsrv/internal/domain/books"

	"microsrv/internal/pkg/mslogger"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type BooksService struct {
	Repo books.BooksRepository
}

var BooksServiceGlobal *BooksService
var onceInitBooksService sync.Once

// InitGlobalBooksService inicializa a variável global para o serviço
func InitBooksService(repo books.BooksRepository) {
	onceInitBooksService.Do(func() {
		BooksServiceGlobal = &BooksService{
			Repo: repo,
		}

		mslogger.LoggerGlobal.Info("Global AutosService configurado com sucesso.")
	})
}

func NewBooksService(
	repo books.BooksRepository,

) *BooksService {
	return &BooksService{
		Repo: repo,
	}
}

func (obj *BooksService) GetBooksModel() (books.BooksRepository, error) {
	if obj.Repo == nil {
		mslogger.LoggerGlobal.Error("BooksService não foi inicializado corretamente: Model é nil.")
		return nil, fmt.Errorf("serviço não inicializado")
	}
	return obj.Repo, nil
}

// ****  INSERT  ******
func (obj *BooksService) Insert(ctx context.Context, body books.BookCreate) (*domain.InsertResult[books.BookID], error) {
	if obj.Repo == nil {
		mslogger.LoggerGlobal.Error("BooksService não foi inicializado corretamente: Model é nil.")
		return nil, fmt.Errorf("serviço não inicializado")
	}
	//Crio o objeto obra e preencho com o corpo da chamada de API
	obra := books.BookCreate{
		NmObra:  body.NmObra,
		NmISBN:  body.NmISBN,
		NmAutor: body.NmAutor,
		NmEdit:  body.NmEdit,
		NrVol:   body.NrVol,
		NrPags:  body.NrPags,
		NrEdi:   body.NrEdi,
		DtEdi:   body.DtEdi,
		DtAqu:   body.DtAqu,
		VrAqu:   body.VrAqu,
		TxtObs:  body.TxtObs,
		SnAtivo: body.SnAtivo,
		UsuInc:  body.UsuInc,
		UsuAlt:  body.UsuInc,
	}

	if obra.SnAtivo == "" {
		obra.SnAtivo = "S"
	}

	now := time.Now()
	if obra.DtInc == nil {
		obra.DtInc = &now
	}
	if obra.DtAlt == nil {
		obra.DtAlt = &now
	}

	res, err := obj.Repo.Insert(ctx, obra)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			mslogger.LoggerGlobal.Error("Erro na inclusão do documento!")
			return nil, err
		}
		return nil, err
	}
	mslogger.LoggerGlobal.Infof("registro inserido com sucesso: %s", res.ID)
	return res, nil
}

func (obj *BooksService) Select(ctx context.Context, id books.BookID) (*books.Book, error) {
	if obj.Repo == nil {
		mslogger.LoggerGlobal.Error("BooksService não foi inicializado corretamente: Model é nil.")
		return nil, fmt.Errorf("serviço não inicializado")
	}

	result, err := obj.Repo.Select(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {

			mslogger.LoggerGlobal.Error("Documento não encontrado!")
			return nil, err
		}
		return nil, err
	}
	mslogger.LoggerGlobal.Infof("registro selecionado com sucesso: %s", id)
	return result, nil
}

func (obj *BooksService) Delete(ctx context.Context, id books.BookID) (*domain.DeleteResult, error) {
	if obj.Repo == nil {
		mslogger.LoggerGlobal.Error("BooksService não foi inicializado corretamente: Model é nil.")
		return nil, fmt.Errorf("serviço não inicializado")
	}

	//result, err := obj.Model.DeleteById(ctx, id)
	result, err := obj.Repo.Delete(ctx, id)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao deletar o registro ID:%s - %v.", id, err)
		return nil, fmt.Errorf("erro ao deletar o registro ID:%s - %v.", id, err)
	}
	if result.DeletedCount == 0 {
		mslogger.LoggerGlobal.Errorf("nenhum registro encontrado para exclusão com ID:%s", id)
		return nil, fmt.Errorf("nenhum registro encontrado para exclusão com ID:%s", id)
	}
	mslogger.LoggerGlobal.Infof("registro deletado com sucesso: %s", id)
	return result, nil
}

func (obj *BooksService) Update(ctx context.Context, id books.BookID, body books.BookUpdate) (*domain.UpdateResult, error) {
	if obj.Repo == nil {
		mslogger.LoggerGlobal.Error("BooksService não foi inicializado corretamente: Model é nil.")
		return nil, fmt.Errorf("serviço não inicializado")
	}
	//Crio o objeto obra e preencho com o corpo da chamada de API
	obra := books.BookUpdate{
		NmObra:  body.NmObra,
		NmISBN:  body.NmISBN,
		NmAutor: body.NmAutor,
		NmEdit:  body.NmEdit,
		NrVol:   body.NrVol,
		NrPags:  body.NrPags,
		NrEdi:   body.NrEdi,
		DtEdi:   body.DtEdi,
		DtAqu:   body.DtAqu,
		VrAqu:   body.VrAqu,
		TxtObs:  body.TxtObs,
		SnAtivo: body.SnAtivo,
		UsuAlt:  body.UsuAlt,
	}

	now := time.Now()
	if obra.DtAlt == nil {
		obra.DtAlt = &now
	}

	//result, err := obj.Model.Update(ctx, id, obra)
	result, err := obj.Repo.Update(ctx, id, obra)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao alterar o registro ID:%s - %v.", id, err)
		return nil, fmt.Errorf("erro ao alterar o registro ID:%s - %v.", id, err)
	}
	if result.MatchedCount == 0 {
		mslogger.LoggerGlobal.Errorf("nenhum documento encontrado para alteração com ID:%s", id)
		return nil, fmt.Errorf("nenhum documento encontrado para alteração com ID:%s", id)
	}
	if result.ModifiedCount == 0 {
		mslogger.LoggerGlobal.Infof("nenhum documento alterado com ID:%s", id)
		return nil, fmt.Errorf("nenhum documento alterado com ID:%s", id)
	}

	mslogger.LoggerGlobal.Infof("registro alterado com sucesso: %s", id)
	return result, nil
}

func (obj *BooksService) SearchByNmObra(ctx context.Context, nmObra string, limit int64) ([]books.Book, error) {
	if obj.Repo == nil {
		mslogger.LoggerGlobal.Error("BooksService não foi inicializado corretamente: Model é nil.")
		return nil, fmt.Errorf("serviço não inicializado")
	}

	result, err := obj.Repo.SearchByNmObra(ctx, nmObra, limit)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao realizar search by:%s - %v.", nmObra, err)
		return nil, fmt.Errorf("Erro ao realizar search by:%s - %v.", nmObra, err)
	}

	mslogger.LoggerGlobal.Infof("search realizado com sucesso: %s", nmObra)
	return result, nil
}
