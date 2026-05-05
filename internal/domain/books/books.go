/*
---------------------------------------------------------------------------------------
File: books.go
Autor: Aldenor
Data: 29-04-2026
----------------------------------------------------------------------------------------
Finalidade:
Entidades do domínio "books"
---------------------------------------------------------------------------------------
*/
package books

import "time"

// BookID é um identificador neutro no domínio.
// Cada implementação concreta converte esse valor para o formato do banco.
type BookID string

type Book struct {
	ID      BookID     `json:"id,omitempty"`
	NmAssu  string     `json:"nm_assu,omitempty"`
	NmISBN  string     `json:"nm_isbn,omitempty"`
	NmObra  string     `json:"nm_obra,omitempty"`
	NmAutor string     `json:"nm_autor,omitempty"`
	NmEdit  string     `json:"nm_edit,omitempty"`
	NrVol   *int       `json:"nr_vol,omitempty"`
	NrPags  *int       `json:"nr_pags,omitempty"`
	NrEdi   *int       `json:"nr_edi,omitempty"`
	DtEdi   *time.Time `json:"dt_edi,omitempty"`
	DtAqu   *time.Time `json:"dt_aqu,omitempty"`
	VrAqu   *float64   `json:"vr_aqu,omitempty"`
	TxtObs  string     `json:"txt_obs,omitempty"`
	SnAtivo string     `json:"sn_ativo,omitempty"`
	DtInc   *time.Time `json:"dt_inc,omitempty"`
	UsuInc  string     `json:"usu_inc,omitempty"`
	DtAlt   *time.Time `json:"dt_alt,omitempty"`
	UsuAlt  string     `json:"usu_alt,omitempty"`
}

// DTO para inserção.
type BookCreate struct {
	NmAssu  string     `json:"nm_assu,omitempty"`
	NmISBN  string     `json:"nm_isbn,omitempty"`
	NmObra  string     `json:"nm_obra,omitempty"`
	NmAutor string     `json:"nm_autor,omitempty"`
	NmEdit  string     `json:"nm_edit,omitempty"`
	NrVol   *int       `json:"nr_vol,omitempty"`
	NrPags  *int       `json:"nr_pags,omitempty"`
	NrEdi   *int       `json:"nr_edi,omitempty"`
	DtEdi   *time.Time `json:"dt_edi,omitempty"`
	DtAqu   *time.Time `json:"dt_aqu,omitempty"`

	VrAqu   *float64   `json:"vr_aqu,omitempty"`
	TxtObs  string     `json:"txt_obs,omitempty"`
	SnAtivo string     `json:"sn_ativo,omitempty"`
	DtInc   *time.Time `json:"dt_inc,omitempty"`
	UsuInc  string     `json:"usu_inc,omitempty"`
	DtAlt   *time.Time `json:"dt_alt,omitempty"`
	UsuAlt  string     `json:"usu_alt,omitempty"`
}

// DTO para atualização parcial.
// Campos com ponteiro permitem distinguir "não informado" de "valor zero".
type BookUpdate struct {
	NmAssu  *string    `json:"nm_assu,omitempty"`
	NmISBN  *string    `json:"nm_isbn,omitempty"`
	NmObra  *string    `json:"nm_obra,omitempty"`
	NmAutor *string    `json:"nm_autor,omitempty"`
	NmEdit  *string    `json:"nm_edit,omitempty"`
	NrVol   *int       `json:"nr_vol,omitempty"`
	NrPags  *int       `json:"nr_pags,omitempty"`
	NrEdi   *int       `json:"nr_edi,omitempty"`
	DtEdi   *time.Time `json:"dt_edi,omitempty"`
	DtAqu   *time.Time `json:"dt_aqu,omitempty"`

	VrAqu   *float64   `json:"vr_aqu,omitempty"`
	TxtObs  *string    `json:"txt_obs,omitempty"`
	SnAtivo *string    `json:"sn_ativo,omitempty"`
	DtAlt   *time.Time `json:"dt_alt,omitempty"`
	UsuAlt  *string    `json:"usu_alt,omitempty"`
}

type BookSearch struct {
	NmSearch string `json:"nm_search,omitempty"`
	NrDocs   int    `json:"nr_docs,omitempty"`
}
