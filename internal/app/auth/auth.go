package auth

import (
	"errors"
	"fmt"
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

const tokenExp = time.Hour * 3
const secretKey = "yandex"

// ChangeID инициализирует и меняет ID для базы данных.
func ChangeID(newID int) {
	ID = newID
	ID++
}

// BuildJWTString создает токен.
func BuildJWTString() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		UserID: ID,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("error from auth - %s", err)
	}

	ID++
	return tokenString, nil
}

// GetUserID проверяет токен.
func GetUserID(tokenString string) (int, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return -1, fmt.Errorf("error from auth - %s", err)
	}

	if !token.Valid {
		return -1, ErrToken
	}

	return claims.UserID, nil
}
