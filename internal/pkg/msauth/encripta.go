/*
---------------------------------------------------------------------------------------
File: encripta.go
Autor: Aldenor
Data: 03-05-2025
Alteração: 04-05-2026
---------------------------------------------------------------------------------------
*/
package auth

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

/*
=========================

	Senhas (bcrypt)

=========================
*/

// HashPassword gera hash bcrypt para persistir no banco.
func HashPassword(password string) (string, error) {
	if strings.TrimSpace(password) == "" {
		return "", errors.New("senha não fornecida")
	}

	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

// CheckPassword compara senha em texto com hash bcrypt.
func CheckPassword(password, hash string) bool {
	if strings.TrimSpace(hash) == "" {
		return false
	}

	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
