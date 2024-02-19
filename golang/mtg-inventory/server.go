package inventory

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/benrm/mtg-inventory/golang/mtg-inventory/scryfall"
)

// Server contains the individual parts of the server
type Server struct {
	Backend Backend

	Server *http.Server

	Scryfall scryfall.Cache
}

// GetUserByUsername returns a user given a username
func (s *Server) GetUserByUsername(rw http.ResponseWriter, req *http.Request) {
	username := req.PathValue("username")

	user, err := s.Backend.GetUserByUsername(req.Context(), username)
	if err != nil {
		if errors.Is(err, ErrUserNoExist) {
			rw.WriteHeader(http.StatusNotFound)
		} else {
			rw.WriteHeader(http.StatusInternalServerError)
		}
		b, err := json.Marshal(HTTPError{Error: err.Error()})
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
