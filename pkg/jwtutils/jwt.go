package jwtutils

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
)

type Claims struct {
	ParticipantID string `json:"participant_id"`
	IsHost        bool   `json:"is_host"`
	jwt.StandardClaims
}

var secretKey = []byte("secret_key")

func GenerateJWT(participantID string, isHost bool) (string, error) {
	expirationTime := time.Now().Add(15 * time.Minute)

	claims := &Claims{
		ParticipantID: participantID,
		IsHost:        isHost,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: &jwt.Time{Time: expirationTime},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(secretKey)

	if err != nil {
		return "", fmt.Errorf("error generating token: %v", err)
	}

	return tokenStr, nil
}

func ValidateToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid or expired token %v", err)
	}

	return claims, nil
}
