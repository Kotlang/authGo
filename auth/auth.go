package auth

import (
	"context"
	"os"

	"github.com/Kotlang/authGo/logger"
	"github.com/dgrijalva/jwt-go"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Claims string

func VerifyToken() grpc_auth.AuthFunc {
	var ACCESS_SECRET = os.Getenv("ACCESS_SECRET")

	return func(ctx context.Context) (context.Context, error) {
		token, err := grpc_auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return nil, err
		}

		parsedToken, err := jwt.ParseWithClaims(
			token,
			&jwt.StandardClaims{},
			func(token *jwt.Token) (interface{}, error) {
				return []byte(ACCESS_SECRET), nil
			})

		claims, ok := parsedToken.Claims.(*jwt.StandardClaims)

		if !ok || !parsedToken.Valid {
			logger.Fatal("Failed validating token", zap.Error(err))
			return nil, status.Errorf(codes.Unauthenticated, "Bad authorization string")
		}

		newCtx := context.WithValue(ctx, Claims("userId"), claims.Id)
		return newCtx, nil
	}
}

func GetToken(tenant, userId, userType string) string {
	atClaims := jwt.StandardClaims{}
	atClaims.Id = userId
	atClaims.Audience = tenant
	atClaims.Subject = userType

	var ACCESS_SECRET = os.Getenv("ACCESS_SECRET")
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, _ := at.SignedString([]byte(ACCESS_SECRET))
	return token
}
