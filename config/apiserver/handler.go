package apiserver

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
)

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (sr SignupRequest) Validate() error {
	if sr.Email == "" {
		return errors.New("email is requered")
	}
	if sr.Password == "" {
		return errors.New("password is requered")
	}
	return nil
}

type Apiresponse[T any] struct {
	Data    *T     `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

func (s *ApiServer) signupHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {
		// var req SignupRequest
		// if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// 	return NewErrorWithStatus(http.StatusBadRequest, fmt.Errorf("invalid rquest body: %v", err))
		// }
		// defer r.Body.Close()

		// if err := req.Validate(); err != nil {
		// 	return NewErrorWithStatus(http.StatusInternalServerError, fmt.Errorf("invalid request: %v", err))
		// }

		req, err := decode[SignupRequest](r)
		if err != nil {
			return NewErrorWithStatus(http.StatusBadRequest, err)
		}

		existingUser, err := s.store.Users.ByEmail(r.Context(), req.Email)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return NewErrorWithStatus(http.StatusInternalServerError, err)
		}

		if existingUser != nil {
			return NewErrorWithStatus(http.StatusConflict, fmt.Errorf("email already registered: %w", err))
		}

		_, err = s.store.Users.CreateUser(r.Context(), req.Email, req.Password)
		if err != nil {
			return NewErrorWithStatus(http.StatusInternalServerError, err)
		}

		if err := encode[Apiresponse[struct{}]](Apiresponse[struct{}]{
			Message: "successfully signed up user",
		}, http.StatusCreated, w); err != nil {
			return NewErrorWithStatus(http.StatusInternalServerError, err)
		}

		return nil
	})
}
