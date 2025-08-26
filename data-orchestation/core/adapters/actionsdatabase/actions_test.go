package actionsdatabase

import (
	"log"
	"testing"
)

func TestCreation(t *testing.T) {

	server := NewActionsDB(
		"localhost",
		"admin",
		"Asdf1234",
		"actions",
		"5432",
	)

	id, err := server.SaveAction("192.168.1.233", 8779.98)

	if err != nil {
		log.Panic(err.Error())
	}

	log.Println(id)
}
