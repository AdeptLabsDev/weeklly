// Package ids gera os identificadores opacos usados em URLs e chaves
// primárias: 80 bits aleatórios em base32 minúscula, 16 caracteres de [a-z2-7].
//
// Opacos de propósito: não revelam ordem, quantidade nem permitem adivinhar
// outros. Curtos o bastante para uma URL legível.
package ids

import (
	"crypto/rand"
	"encoding/base32"
	"strings"
)

// Length é o tamanho fixo de um identificador.
const Length = 16

var encoding = base32.StdEncoding.WithPadding(base32.NoPadding)

// New gera um identificador novo.
func New() string {
	var b [10]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err) // crypto/rand só falha se o sistema estiver sem entropia; não há o que fazer
	}
	return strings.ToLower(encoding.EncodeToString(b[:]))
}

// Valid informa se s tem o formato gerado por New. Serve para rejeitar URLs
// malformadas antes de consultar o banco.
func Valid(s string) bool {
	if len(s) != Length {
		return false
	}
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < '2' || r > '7') {
			return false
		}
	}
	return true
}
