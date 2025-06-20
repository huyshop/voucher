package utils

import (
	"log"

	gonanoid "github.com/matoous/go-nanoid"
	"github.com/rs/xid"
)

const (
	CharsetUpper string = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func MakeVoucherId() string {
	return "vou" + xid.New().String()
}

func MakeCodeId() string {
	return "cod" + xid.New().String()
}

func MakeUserVoucherId() string {
	return "usv" + xid.New().String()
}

func MakeCode() string {
	id, err := gonanoid.Generate(CharsetUpper, 9)
	if err != nil {
		log.Print(err)
		id, _ := gonanoid.Generate(CharsetUpper, 9)
		return id
	}
	return id
}
