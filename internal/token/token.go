package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

// JWT - структура для создания JWT токены
type JWT struct {
	secretKey []byte
}

func NewJWT(secretKey string) (*JWT, error) {
	if len(secretKey) == 0 {
		return nil, fmt.Errorf("empty secret key")
	}
	return &JWT{
		secretKey: []byte(secretKey),
	}, nil
}

// JWTClaims описание записей в токене JWT
type JWTClaims struct {
	jwt.StandardClaims
}

// JWTExpire - время жизни токена
const JWTExpire = time.Hour * 3

// BuildJWT - метод для формирования JWT токена с добавлением UUID пользователя
func (j *JWT) BuildJWT(userID string) (string, error) {
	now := time.Now()
	exp := now.Add(JWTExpire)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, JWTClaims{
		StandardClaims: jwt.StandardClaims{
			Id:        userID,
			ExpiresAt: exp.Unix(),
			IssuedAt:  now.Unix(),
			NotBefore: now.Unix(),
		},
	})

	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseJWT - метод разбора JWT токена с проверкой секрета и возвратом кастомных записей
func (j *JWT) ParseJWT(token string) (*JWTClaims, error) {
	claims := &JWTClaims{}

	jwtToken, err := jwt.ParseWithClaims(token, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return j.secretKey, nil
		})
	if err != nil {
		return nil, err
	}
	if !jwtToken.Valid {
		return nil, fmt.Errorf("token is not valid")
	}
	return claims, nil
}

// DecodeUserId - метод извлечения ID пользователя из токена
func (j *JWT) DecodeUserId(token string) (string, error) {
	claims, err := j.ParseJWT(token)
	if err != nil {
		return "", err
	}
	return claims.StandardClaims.Id, nil
}
