package inventory

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/benrm/mtg-inventory/golang/mtg-inventory/scryfall"
	intsql "github.com/benrm/mtg-inventory/golang/mtg-inventory/sql"
)

// Server contains the individual parts of the server
type Server struct {
	DB *sql.DB

	Server *http.Server

	Scryfall scryfall.Cache
}

// Error is the type used to marshal errors into JSON
type Error struct {
	Error string `json:"error"`
}

// GetUserByUsername returns a user given a username
func (s *Server) GetUserByUsername(rw http.ResponseWriter, req *http.Request) {
	username := req.PathValue("username")

	user, err := intsql.GetUserByUsername(req.Context(), s.DB, username)
	if err != nil {
		if errors.Is(err, intsql.ErrUserNoExist) {
			rw.WriteHeader(http.StatusNotFound)
		} else {
			rw.WriteHeader(http.StatusInternalServerError)
		}
		b, err := json.Marshal(Error{Error: err.Error()})
		if err != nil {
			log.Printf("Error marshaling error in GetUserByUsername: %s", err.Error())
			return
		}
		_, err = rw.Write(b)
		if err != nil {
			log.Printf("Error writing error in GetUserByUsername: %s", err.Error())
		}
		return
	}

	rw.WriteHeader(http.StatusOK)
	b, err := json.Marshal(user)
	if err != nil {
		log.Printf("Error marshaling user in GetUserByUsername: %s", err.Error())
		return
	}
	_, err = rw.Write(b)
	if err != nil {
		log.Printf("Error writing user in GetUserByUsername: %s", err.Error())
	}
}
