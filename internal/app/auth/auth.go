package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

var (
	ID       = 1
	ErrToken = errors.New("token is not valid")
)

const TOKEN_EXP = time.Hour * 3
const SECRET_KEY = "yandex"

// BuildJWTString создает токен
func BuildJWTString() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP)),
		},
		UserID: ID,
	})

	tokenString, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", err
	}

	ID++
	return tokenString, nil
}

// GetUserID проверяет токен
func GetUserID(tokenString string) (int, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})
	if err != nil {
		return -1, errors.New("something wrong")
	}

	if !token.Valid {
		return -1, ErrToken
	}

	return claims.UserID, nil
}
