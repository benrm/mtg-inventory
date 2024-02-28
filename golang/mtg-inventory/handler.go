package inventory

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/benrm/mtg-inventory/golang/mtg-inventory/scryfall"
)

// Handler contains the individual parts of the server
type Handler struct {
	Backend Backend

	Scryfall scryfall.Cache

	*http.ServeMux
}

// NewHandler returns a new Handler
func NewHandler(backend Backend, scryfallCache scryfall.Cache) *Handler {
	handler := &Handler{
		Backend:  backend,
		Scryfall: scryfallCache,
		ServeMux: http.NewServeMux(),
	}

	handler.ServeMux.HandleFunc("GET /user/{username}", handler.GetUser)
	handler.ServeMux.HandleFunc("POST /user", handler.PostUser)

	return handler
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

// GetUser returns a user given a username
func (s *Handler) GetUser(rw http.ResponseWriter, req *http.Request) {
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
			log.Printf("Error in GetUser: %s", err.Error())
		}
		return
	}

	rw.WriteHeader(http.StatusOK)
	err = writeJSON(rw, user)
	if err != nil {
		log.Printf("Error in GetUser: %s", err.Error())
	}
}

// PostUser creates a user given a username and an email
func (s *Handler) PostUser(rw http.ResponseWriter, req *http.Request) {
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

	user, err := s.Backend.AddUser(req.Context(), input.Username)
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
