package inventory

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/benrm/mtg-inventory/golang/mtg-inventory/scryfall"
)

// Server contains the individual parts of the server
type Server struct {
	Backend Backend

	Server *http.Server

	Scryfall scryfall.Cache
}

func writeJSON(rw io.Writer, object interface{}) error {
	b, err := json.Marshal(object)
	if err != nil {
		return fmt.Errorf("error marshaling %T: %w", object, err)
	}
	_, err = rw.Write(b)
	if err != nil {
		return fmt.Errorf("error writing %T: %w", object, err)
	}
	return nil
}

// GetUserByID returns a user given an ID
func (s *Server) GetUserByID(rw http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		err = writeJSON(rw, HTTPError{Error: err.Error()})
		if err != nil {
			log.Printf("Error in GetUserByID: %s", err.Error())
		}
		return
	}

	_, _ = io.ReadAll(req.Body)

	user, err := s.Backend.GetUserByID(req.Context(), id)
	if err != nil {
		if errors.Is(err, ErrUserNoExist) {
			rw.WriteHeader(http.StatusNotFound)
		} else {
			rw.WriteHeader(http.StatusInternalServerError)
		}
		err = writeJSON(rw, HTTPError{Error: err.Error()})
		if err != nil {
			log.Printf("Error in GetUserByID: %s", err.Error())
		}
		return
	}

	rw.WriteHeader(http.StatusOK)
	err = writeJSON(rw, user)
	if err != nil {
		log.Printf("Error in GetUserByID: %s", err.Error())
	}
}

// GetUserByUsername returns a user given a username
func (s *Server) GetUserByUsername(rw http.ResponseWriter, req *http.Request) {
	username := req.PathValue("username")

	_, _ = io.ReadAll(req.Body)

	user, err := s.Backend.GetUserByUsername(req.Context(), username)
	if err != nil {
		if errors.Is(err, ErrUserNoExist) {
			rw.WriteHeader(http.StatusNotFound)
		} else {
			rw.WriteHeader(http.StatusInternalServerError)
		}
		err = writeJSON(rw, HTTPError{Error: err.Error()})
		if err != nil {
			log.Printf("Error in GetUserByUsername: %s", err.Error())
		}
		return
	}

	rw.WriteHeader(http.StatusOK)
	err = writeJSON(rw, user)
	if err != nil {
		log.Printf("Error in GetUserByUsername: %s", err.Error())
	}
}

// PostUser creates a user given a username and an email
func (s *Server) PostUser(rw http.ResponseWriter, req *http.Request) {
	if req.Header.Get("Content-Type") != "application/json" {
		rw.WriteHeader(http.StatusUnsupportedMediaType)
		err := writeJSON(rw, HTTPError{Error: "must use 'application/json'"})
		if err != nil {
			log.Printf("Error in PostUser around unsupported media type: %s", err.Error())
		}
		return
	}

	b, err := io.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		err = writeJSON(rw, HTTPError{Error: err.Error()})
		if err != nil {
			log.Printf("Error in PostUser around reading body: %s", err.Error())
		}
		return
	}

	input := struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}{}
	err = json.Unmarshal(b, &input)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		err = writeJSON(rw, HTTPError{Error: err.Error()})
		if err != nil {
			log.Printf("Error in PostUser around unmarshal JSON: %s", err.Error())
		}
		return
	}

	user, err := s.Backend.AddUser(req.Context(), input.Username, input.Email)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		err = writeJSON(rw, HTTPError{Error: err.Error()})
		if err != nil {
			log.Printf("Error in PostUser around adding user: %s", err.Error())
		}
		return
	}

	err = writeJSON(rw, user)
	if err != nil {
		log.Printf("Error in PostUser around writing user: %s", err.Error())
	}
}
