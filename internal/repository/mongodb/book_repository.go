/*
---------------------------------------------------------------------------------------
File: books_repository.go
Autor: Aldenor
Data: 18-04-2026
Finalidade:
Tipo concreto que realiza as operações CRUD no banco de dados MongoDB, relacionadas à
collection Books. Ela atende aos requisitos da interface Repository, que generaliza as
operações CRUD no banco de dados.
---------------------------------------------------------------------------------------
*/
package mongodb

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"microsrv/internal/domain"
	"microsrv/internal/domain/books"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type BooksMongoRepository struct {
	Collection *mongo.Collection
}

func NewBooksMongoRepository(collection *mongo.Collection) *BooksMongoRepository {
	return &BooksMongoRepository{
		Collection: collection,
	}
}

//var _ BooksMongoRepository = (*BooksMongoRepository)(nil)

type bookDocument struct {
	ID      bson.ObjectID `bson:"_id,omitempty"`
	NmAssu  string        `bson:"nm_assu,omitempty"`
	NmISBN  string        `bson:"nm_isbn,omitempty"`
	NmObra  string        `bson:"nm_obra,omitempty"`
	NmAutor string        `bson:"nm_autor,omitempty"`
	NmEdit  string        `bson:"nm_edit,omitempty"`
	NrVol   *int          `bson:"nr_vol,omitempty"`
	NrPags  *int          `bson:"nr_pags,omitempty"`
	NrEdi   *int          `bson:"nr_edi,omitempty"`
	DtEdi   *time.Time    `bson:"dt_edi,omitempty"`
	DtAqu   *time.Time    `bson:"dt_aqu,omitempty"`
	VrAqu   *float64      `bson:"vr_aqu,omitempty"`
	TxtObs  string        `bson:"txt_obs,omitempty"`
	SnAtivo string        `bson:"sn_ativo,omitempty"`
	DtInc   *time.Time    `bson:"dt_inc,omitempty"`
	UsuInc  string        `bson:"usu_inc,omitempty"`
	DtAlt   *time.Time    `bson:"dt_alt,omitempty"`
	UsuAlt  string        `bson:"usu_alt,omitempty"`
}

func toMongoDocumentCreate(src books.BookCreate) bookDocument {
	now := time.Now()

	return bookDocument{
		NmAssu:  src.NmAssu,
		NmISBN:  src.NmISBN,
		NmObra:  src.NmObra,
		NmAutor: src.NmAutor,
		NmEdit:  src.NmEdit,
		NrVol:   src.NrVol,
		NrPags:  src.NrPags,
		NrEdi:   src.NrEdi,
		DtEdi:   src.DtEdi,
		DtAqu:   src.DtAqu,
		VrAqu:   src.VrAqu,
		TxtObs:  src.TxtObs,
		SnAtivo: src.SnAtivo,
		DtInc:   &now,
		UsuInc:  src.UsuInc,
	}
}

func toBookModel(src bookDocument) books.Book {
	return books.Book{
		ID:      books.BookID(src.ID.Hex()),
		NmAssu:  src.NmAssu,
		NmISBN:  src.NmISBN,
		NmObra:  src.NmObra,
		NmAutor: src.NmAutor,
		NmEdit:  src.NmEdit,
		NrVol:   src.NrVol,
		NrPags:  src.NrPags,
		NrEdi:   src.NrEdi,
		DtEdi:   src.DtEdi,
		DtAqu:   src.DtAqu,
		VrAqu:   src.VrAqu,
		TxtObs:  src.TxtObs,
		SnAtivo: src.SnAtivo,
		DtInc:   src.DtInc,
		UsuInc:  src.UsuInc,
		DtAlt:   src.DtAlt,
		UsuAlt:  src.UsuAlt,
	}
}

func buildBookUpdate(entity books.BookUpdate) (bson.M, error) {
	setFields := bson.M{}

	if entity.NmAssu != nil {
		setFields["nm_assu"] = *entity.NmAssu
	}
	if entity.NmISBN != nil {
		setFields["nm_isbn"] = *entity.NmISBN
	}
	if entity.NmObra != nil {
		setFields["nm_obra"] = *entity.NmObra
	}
	if entity.NmAutor != nil {
		setFields["nm_autor"] = *entity.NmAutor
	}
	if entity.NmEdit != nil {
		setFields["nm_edit"] = *entity.NmEdit
	}
	if entity.NrVol != nil {
		setFields["nr_vol"] = *entity.NrVol
	}
	if entity.NrPags != nil {
		setFields["nr_pags"] = *entity.NrPags
	}
	if entity.NrEdi != nil {
		setFields["nr_edi"] = *entity.NrEdi
	}
	if entity.DtEdi != nil {
		setFields["dt_edi"] = entity.DtEdi
	}
	if entity.DtAqu != nil {
		setFields["dt_aqu"] = entity.DtAqu
	}
	if entity.VrAqu != nil {
		setFields["vr_aqu"] = *entity.VrAqu
	}
	if entity.TxtObs != nil {
		setFields["txt_obs"] = *entity.TxtObs
	}
	if entity.SnAtivo != nil {
		setFields["sn_ativo"] = *entity.SnAtivo
	}
	if entity.UsuAlt != nil {
		setFields["usu_alt"] = *entity.UsuAlt
	}

	now := time.Now()
	setFields["dt_alt"] = &now

	if len(setFields) == 0 {
		return nil, fmt.Errorf("nenhum campo informado para atualização")
	}

	return setFields, nil
}

func parseObjectID(id books.BookID) (bson.ObjectID, error) {
	objID, err := bson.ObjectIDFromHex(string(id))
	if err != nil {
		return bson.ObjectID{}, fmt.Errorf("id inválido para MongoDB: %w", err)
	}
	return objID, nil
}

func (r *BooksMongoRepository) Insert(ctx context.Context, entity books.BookCreate) (*domain.InsertResult[books.BookID], error) {
	if r.Collection == nil {
		return nil, fmt.Errorf("coleção do MongoDB não foi inicializada")
	}

	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	doc := toMongoDocumentCreate(entity)

	res, err := r.Collection.InsertOne(cctx, doc)
	if err != nil {
		return nil, fmt.Errorf("erro ao inserir documento: %w", err)
	}

	insertedID, ok := res.InsertedID.(bson.ObjectID)
	if !ok {
		return nil, fmt.Errorf("não foi possível converter o id inserido para ObjectID")
	}

	return &domain.InsertResult[books.BookID]{
		ID: books.BookID(insertedID.Hex()),
	}, nil
}

func (r *BooksMongoRepository) Select(ctx context.Context, id books.BookID) (*books.Book, error) {
	if r.Collection == nil {
		return nil, fmt.Errorf("coleção do MongoDB não foi inicializada")
	}

	objID, err := parseObjectID(id)
	if err != nil {
		return nil, err
	}

	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.D{{Key: "_id", Value: objID}}

	var result bookDocument
	err = r.Collection.FindOne(cctx, filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		return nil, fmt.Errorf("erro ao buscar documento por id: %w", err)
	}

	book := toBookModel(result)
	return &book, nil
}

func (r *BooksMongoRepository) Update(ctx context.Context, id books.BookID, entity books.BookUpdate) (*domain.UpdateResult, error) {
	if r.Collection == nil {
		return nil, fmt.Errorf("coleção do MongoDB não foi inicializada")
	}

	uCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	setFields, err := buildBookUpdate(entity)
	if err != nil {
		return nil, err
	}

	// Chave de busca/filtro
	objID, err := parseObjectID(id)
	if err != nil {
		return nil, err
	}
	filter := bson.D{{Key: "_id", Value: objID}}

	// Dados a serem modificados
	update := bson.D{{Key: "$set", Value: setFields}}

	// Impede a criação de novo documento
	opts := options.UpdateOne().SetUpsert(false)

	//Executa a chamda ao engine
	res, err := r.Collection.UpdateOne(uCtx, filter, update, opts)
	if err != nil {
		return nil, fmt.Errorf("erro ao alterar documento: %w", err)
	}

	//mslogger.LoggerGlobal.Infof("%w", res)

	return &domain.UpdateResult{
		MatchedCount:  res.MatchedCount,
		ModifiedCount: res.ModifiedCount,
	}, nil
}

func (r *BooksMongoRepository) Delete(ctx context.Context, id books.BookID) (*domain.DeleteResult, error) {
	if r.Collection == nil {
		return nil, fmt.Errorf("coleção do MongoDB não foi inicializada")
	}

	objID, err := parseObjectID(id)
	if err != nil {
		return nil, err
	}

	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.D{{Key: "_id", Value: objID}}

	res, err := r.Collection.DeleteOne(cctx, filter)
	if err != nil {
		return nil, fmt.Errorf("erro ao excluir documento: %w", err)
	}

	return &domain.DeleteResult{
		DeletedCount: res.DeletedCount,
	}, nil
}

func buildNmObraApproxFilter(nmObra string) (bson.D, error) {
	termo := strings.TrimSpace(nmObra)
	if termo == "" {
		return nil, fmt.Errorf("nome da obra não informado")
	}

	partes := strings.Fields(termo)
	if len(partes) == 0 {
		return nil, fmt.Errorf("nome da obra não informado")
	}

	// Se houver apenas um termo, faz busca simples
	if len(partes) == 1 {
		return bson.D{
			{Key: "nm_obra", Value: bson.Regex{
				Pattern: regexp.QuoteMeta(partes[0]),
				Options: "i",
			}},
		}, nil
	}

	// Se houver vários termos, exige que todos apareçam no campo
	conds := bson.A{}
	for _, p := range partes {
		conds = append(conds, bson.D{
			{Key: "nm_obra", Value: bson.Regex{
				Pattern: regexp.QuoteMeta(p),
				Options: "i",
			}},
		})
	}

	return bson.D{
		{Key: "$and", Value: conds},
	}, nil
}

func (r *BooksMongoRepository) SearchByNmObra(ctx context.Context, nmObra string, limit int64) ([]books.Book, error) {
	if r.Collection == nil {
		return nil, fmt.Errorf("coleção do MongoDB não foi inicializada")
	}

	if limit <= 0 {
		limit = 20
	}

	filter, err := buildNmObraApproxFilter(nmObra)
	if err != nil {
		return nil, err
	}

	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	opts := options.Find().
		SetLimit(limit).
		SetSort(bson.D{{Key: "nm_obra", Value: 1}})

	cursor, err := r.Collection.Find(cctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar livros por nome da obra: %w", err)
	}
	defer cursor.Close(cctx)

	var docs []bookDocument
	if err := cursor.All(cctx, &docs); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resultado da busca: %w", err)
	}

	books := make([]books.Book, 0, len(docs))
	for _, doc := range docs {
		books = append(books, toBookModel(doc))
	}

	return books, nil
}
