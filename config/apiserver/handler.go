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
		req, err := decode[SignupRequest](r)
		if err != nil {
			return NewErrorWithStatus(http.StatusBadRequest, err)
		}

		existingUser, err := s.store.Users.ByEmail(r.Context(), req.Email)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return NewErrorWithStatus(http.StatusInternalServerError, err)
		}

		if existingUser != nil {
			return NewErrorWithStatus(http.StatusConflict, fmt.Errorf("email already registered"))
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

type SigninRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SigninResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (sr *SigninRequest) Validate() error {
	if sr.Email == "" {
		return errors.New("email is required")
	}
	if sr.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

func (s *ApiServer) signinHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {
		req, err := decode[*SigninRequest](r)
		if err != nil {
			return NewErrorWithStatus(http.StatusBadRequest, err)
		}

		user, err := s.store.Users.ByEmail(r.Context(), req.Email)
		if err != nil {
			return NewErrorWithStatus(http.StatusInternalServerError, err)
		}

		if err := user.ComparePassword(req.Password); err != nil {
			return NewErrorWithStatus(http.StatusUnauthorized, err)
		}

		tokenPair, err := s.jwtManager.GenerateTokenPair(user.Id)
		if err != nil {
			return NewErrorWithStatus(http.StatusInternalServerError, err)
		}

		if _, err = s.store.RefreshTokenStore.Delete(r.Context(), user.Id); err != nil {
			return NewErrorWithStatus(http.StatusInternalServerError, err)
		}

		_, err = s.store.RefreshTokenStore.Create(r.Context(), user.Id, tokenPair.RefreshToken)
		if err != nil {
			return NewErrorWithStatus(http.StatusInternalServerError, err)
		}

		if err := encode(Apiresponse[SigninResponse]{
			Data: &SigninResponse{
				AccessToken:  tokenPair.AccessToken.Raw,
				RefreshToken: tokenPair.RefreshToken.Raw,
			},
		}, http.StatusOK, w); err != nil {
			return NewErrorWithStatus(http.StatusInternalServerError, err)
		}

		return nil
	})
}
