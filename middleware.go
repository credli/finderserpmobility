package main

import (
	"errors"
	"github.com/gorilla/context"
	"log"
	"net/http"
	"strings"
)

type contextKey int

const USERDATA contextKey = 0

func (o *OAuthHandler) MiddlewareFunc(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			o.unauthorized(w)
			return
		}

		token, err := o.extractToken(authHeader)
		if err != nil {
			http.Error(w, deferror.Get(E_MISSING_AUTHENTICATION), http.StatusBadRequest)
			return
		}

		accessData, err := o.server.Storage.LoadAccess(token)
		if err != nil {
			log.Printf("Error: %s\n", err)
			http.Error(w, deferror.Get(E_INVALID_AUTHENTICATION), http.StatusBadRequest)
			return
		}

		if accessData.Client == nil {
			o.unauthorized(w)
			return
		}

		if accessData.Client.GetRedirectUri() == "" {
			o.unauthorized(w)
			return
		}

		if accessData.IsExpired() {
			o.unauthorized(w)
			return
		}

		context.Set(r, USERDATA, accessData.UserData)

		handler(w, r)
	}
}

func (o *OAuthHandler) extractToken(header string) (string, error) {
	parts := strings.SplitN(header, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		return "", errors.New(deferror.Get(E_INVALID_AUTHENTICATION))
	}
	return parts[1], nil
}

func (o *OAuthHandler) unauthorized(w http.ResponseWriter) {
	http.Error(w, deferror.Get(E_UNAUTHORIZED), http.StatusUnauthorized)
}
