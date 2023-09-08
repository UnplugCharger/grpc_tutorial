package service

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"time"
)

var ErrInvalidAccessToken = errors.New("invalid access token")

type JwtManager struct {
	secretKey     string
	tokenDuration time.Duration
}

type UserClaim struct {
	jwt.StandardClaims
	Username string `json:"username"`
	Role     string `json:"role"`
}

func NewJwtManager(secretKey string, tokenDuration time.Duration) *JwtManager {
	return &JwtManager{secretKey: secretKey, tokenDuration: tokenDuration}
}

func (manager *JwtManager) Generate(user *User) (string, error) {
	claim := UserClaim{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(manager.tokenDuration).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		Username: user.Username,
		Role:     user.Role,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString([]byte(manager.secretKey))
}

func (manager *JwtManager) Verify(accessToken string) (*UserClaim, error) {
	token, err := jwt.ParseWithClaims(accessToken, &UserClaim{}, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok { // if not ok, return error
			return nil, ErrInvalidAccessToken
		}
		return []byte(manager.secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	claim, ok := token.Claims.(*UserClaim)
	if !ok {
		return nil, ErrInvalidAccessToken
	}
	return claim, nil
}
