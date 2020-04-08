package lib

import (
	"encoding/json"
	"fmt"
	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/render"
	"net/http"
)

type Reponse struct {
	Message string `json:"message"`
}

type Response struct {
	Message string `json:"message"`
}

type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

type JSONWebKeys struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

func getPemCert(token *jwt.Token) (string, error) {
	cert := ""
	resp, err := http.Get("https://dohrm.eu.auth0.com/.well-known/jwks.json")

	if err != nil {
		return cert, err
	}
	defer func() { _ = resp.Body.Close() }()

	var jwks = Jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)

	if err != nil {
		return cert, err
	}

	for k, _ := range jwks.Keys {
		if token.Header["kid"] == jwks.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		err := HttpUnauthorized(fmt.Errorf("unable to find appropriate key"))
		return cert, err
	}

	return cert, nil
}

func AuthHttpMiddleware(auth0Client string, auth0Secret string, contextKey string) func(http.Handler) http.Handler {
	res := jwtmiddleware.New(jwtmiddleware.Options{
		UserProperty: contextKey,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err string) {
			_ = render.Render(w, r, HttpUnauthorized(fmt.Errorf(err)))
		},
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			aud := auth0Client
			checkAud := token.Claims.(jwt.MapClaims).VerifyAudience(aud, false)
			if !checkAud {
				return token, HttpUnauthorized(fmt.Errorf("invalid audience"))
			}
			// Verify 'iss' claim
			iss := "https://dohrm.eu.auth0.com/"
			checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, false)
			if !checkIss {
				return token, HttpUnauthorized(fmt.Errorf("invalid issuer"))
			}

			cert, err := getPemCert(token)
			if err != nil {
				return nil, err
			}

			result, err := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
			if err != nil {
				return nil, HttpUnauthorized(err)
			}
			return result, nil
		},
	})
	return func(next http.Handler) http.Handler {
		return res.Handler(next)
	}
}
