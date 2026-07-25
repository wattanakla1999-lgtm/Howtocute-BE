package service

import (
	"errors"
	"fmt"
	"nailly-back-end/internal/model"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const CustomerTokenType = "customer"

type CustomerClaims struct {
	Type       string `json:"typ"`
	LineUserID string `json:"lineUserId"`
	jwt.RegisteredClaims
}

type CustomerJWTManager struct {
	secret []byte
	ttl    time.Duration
}

func NewCustomerJWTManager(secret string, ttl time.Duration) *CustomerJWTManager {
	return &CustomerJWTManager{secret: []byte(secret), ttl: ttl}
}

func (m *CustomerJWTManager) Generate(customer model.User) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(m.ttl)
	lineUserID := ""
	if customer.LineUserID != nil {
		lineUserID = *customer.LineUserID
	}
	claims := CustomerClaims{
		Type:       CustomerTokenType,
		LineUserID: lineUserID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jwtIssuer,
			Subject:   strconv.FormatUint(uint64(customer.ID), 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	return token, expiresAt, err
}

func (m *CustomerJWTManager) Verify(tokenValue string) (*CustomerClaims, error) {
	claims := &CustomerClaims{}
	token, err := jwt.ParseWithClaims(
		tokenValue,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
			}
			return m.secret, nil
		},
		jwt.WithIssuer(jwtIssuer),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}
	if claims.Type != CustomerTokenType || claims.Subject == "" || claims.LineUserID == "" {
		return nil, errors.New("invalid customer token")
	}
	return claims, nil
}
