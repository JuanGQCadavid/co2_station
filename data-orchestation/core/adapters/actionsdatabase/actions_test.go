package actionsdatabase

import "testing"

func TestCreation(t *testing.T) {

	server := NewActionsDB(
		"localhost",
		"admin",
		"Asdf1234",
		"actions",
		"5432",
	)

	server.SaveAction("192.168.1.233", 8779.98)
}
