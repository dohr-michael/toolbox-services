package app_context

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
)

const (
	AuthTokenContextKey = "ctx:auth:token"
)

func GetAuthToken(ctx context.Context) *jwt.Token {
	r, ok := ctx.Value(AuthTokenContextKey).(*jwt.Token)
	if !ok {
		panic(fmt.Sprintf("cannot found key Nats in context"))
	}
	return r
}

func GetAuthUser(ctx context.Context) string {
	token := GetAuthToken(ctx)
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		panic(fmt.Sprintf("cannot read claims of context"))
	}
	return claims["email"].(string)
}
