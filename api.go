package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type ApiServer struct {
  store Storage
}

func NewApiServer(s Storage) *ApiServer {
  return &ApiServer{
    store: s,
  }
}

func (s *ApiServer) HandleRegister(w http.ResponseWriter, r *http.Request) {
  // get register request
  registerReq := new(UserJSON)
  err := json.NewDecoder(r.Body).Decode(registerReq)
  if err != nil {
    err = fmt.Errorf("handle register: json decode error: %w", err)
    slog.Error(err.Error())
    http.Error(w, "internal server error", http.StatusInternalServerError)
    return
  }

  // check for existing user
  _, err = s.store.GetUserByUsername(registerReq.Username)
  if err == nil {
    err = fmt.Errorf("handle register: username %s already exists", registerReq.Username) 
    slog.Error(err.Error())
    err = fmt.Errorf("username %s is taken", registerReq.Username)
    http.Error(w, err.Error(), http.StatusBadRequest)
  }

  // encrypt password
  encPass, err := bcrypt.GenerateFromPassword([]byte(registerReq.PasswordHashed), bcrypt.DefaultCost)
  if err != nil {
    err = fmt.Errorf("handle register: bcrypt password gen error: %w", err)
    slog.Error(err.Error())
    http.Error(w, "internal server error", http.StatusInternalServerError)
    return
  }
  registerReq.PasswordHashed = string(encPass)

  // create user
  user, err := s.store.CreateUser(*registerReq)
  if err != nil {
    err = fmt.Errorf("handle register: create user failed: %w", err)
    slog.Error(err.Error())
    http.Error(w, "internal server error", http.StatusInternalServerError)
    return
  }

  // create access token
  tokenString, err := CreateJWT(user.Id)
  if err != nil {
    err = fmt.Errorf("handle register: create jwt error: %w", err)
    slog.Error(err.Error())
    http.Error(w, "internal server error", http.StatusInternalServerError)
    return
  }

  // create cookie
  cookie := CreateCookie("accessToken", tokenString) 
  http.SetCookie(w, cookie)

  // response
  if err = WriteJSON(w, user, http.StatusOK); err != nil {
    err = fmt.Errorf("handle register: write json error: %w", err)
    slog.Error(err.Error())
    http.Error(w, "internal server error", http.StatusInternalServerError)
    return
  }
}

func (s *ApiServer) HandleLogin(w http.ResponseWriter, r *http.Request) {
  // get login request
  loginReq := new(UserJSON)
  err := json.NewDecoder(r.Body).Decode(loginReq)
  if err != nil {
    err = fmt.Errorf("handle login: json decode error: %w", err)
    slog.Error(err.Error())
    http.Error(w, "internal server error", http.StatusInternalServerError)
    return
  }

  // get user
  user, err := s.store.GetUserByUsername(loginReq.Username)
  if err != nil {
    err = fmt.Errorf("handle login: user %s not found", loginReq.Username)
    slog.Error(err.Error())
    err = fmt.Errorf("user %s not found", loginReq.Username)
    http.Error(w, err.Error(), http.StatusNotFound)
    return
  }

  // validate password
  if !user.ValidatePassword(loginReq.PasswordHashed) {
    err = fmt.Errorf("handle login: wrong password")
    slog.Error(err.Error())
    http.Error(w, "wrong password", http.StatusBadRequest)
    return
  }

  // create access token
  token, err := CreateJWT(user.Id)
  if err != nil {
    err = fmt.Errorf("handle login: create jwt error: %w", err)
    slog.Error(err.Error())
    http.Error(w, "internal server error", http.StatusInternalServerError)
    return
  }

  // create cookie
  cookie := CreateCookie("accessToken", token)
  http.SetCookie(w, cookie)

  // response
  if err = WriteJSON(w, user, http.StatusOK); err != nil {
    err = fmt.Errorf("handle login: write json error: %w", err)
    slog.Error(err.Error())
    http.Error(w, "internal server error", http.StatusInternalServerError)
    return
  }
}

func (s *ApiServer) ProtectMiddleware(next http.HandlerFunc) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    // get cookie
    cookie, err := r.Cookie("accessToken")
    if err != nil {
      err = fmt.Errorf("protect middleware: failed to get cookie: %w", err)
      slog.Error(err.Error())
      http.Error(w, "not authorized", http.StatusUnauthorized)
      return
    }

    // validate token
    token, err := ValidateJWT(cookie.Value)
    if err != nil {
      err = fmt.Errorf("protect middleware: failed to validate jwt: %w", err)
      slog.Error(err.Error())
      http.Error(w, "not authorized", http.StatusUnauthorized)
      return
    }

    // get claims
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
      err = fmt.Errorf("protect middleware: claims type assertion failed: %w", err)
      slog.Error(err.Error())
      http.Error(w, "not authorized", http.StatusUnauthorized)
      return
    }

    // get userId of type STRING
    userIdStr, ok := claims["userId"].(string)
    if !ok {
      err = fmt.Errorf("protect middleware: failed to get userId from claims: %w", err)
      slog.Error(err.Error())
      http.Error(w, "not authorized", http.StatusUnauthorized)
      return
    }

    // conversion
    userId, err := strconv.Atoi(userIdStr)
    if err != nil {
      err = fmt.Errorf("protect middleware: failed userId string to int conversion: %w", err)
      slog.Error(err.Error())
      http.Error(w, "not authorized", http.StatusUnauthorized)
      return
    }

    // get user from store
    user, err := s.store.GetUserById(userId)
    if err != nil {
      err = fmt.Errorf("protect middleware: failed to get user with id(%d): %w", userId, err)
      slog.Error(err.Error())
      http.Error(w, "not authorized", http.StatusUnauthorized)
      return
    }

    // add user to r.context
    ctx := context.WithValue(r.Context(), ContextKeyUser, user)
    next(w, r.WithContext(ctx))
  }
}
