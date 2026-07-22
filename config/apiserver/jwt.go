package apiserver

import (
	"fmt"
	"go_sqs_pqsql_s3_project/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var signingMethod = jwt.SigningMethodHS256

type JwtManager struct {
	Config *config.Config
}

func NewJwtManger(config *config.Config) *JwtManager {
	return &JwtManager{
		Config: config,
	}
}

type TokenPair struct {
	AccessToken  *jwt.Token
	RefreshToken *jwt.Token
}

type CustomClaims struct {
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func (jm *JwtManager) Parse(token string) (*jwt.Token, error) {
	parser := jwt.NewParser()
	jwtToken, err := parser.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if t.Method != signingMethod {
			return nil, fmt.Errorf("unexpected sign method: %v", t.Header["alg"])
		}
		return []byte(jm.Config.JwtSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	return jwtToken, nil
}

func (jm *JwtManager) IsAccessToken(token *jwt.Token) bool {
	jwtClaims := token.Claims.(jwt.MapClaims)
	if tokenType, ok := jwtClaims["token_type"]; ok {
		return tokenType == "access"
	}
	return false
}

func (jm *JwtManager) GenerateTokenPair(userId uuid.UUID) (*TokenPair, error) {
	now := time.Now()
	issuer := "http://" + jm.Config.ApiServerHost + ":" + jm.Config.ApiServerPort
	accessJwtToken, err := jm.getJwtToken("access", 15*time.Minute, now, issuer, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshJwtToken, err := jm.getJwtToken("refresh", 24*time.Hour*30, now, issuer, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessJwtToken,
		RefreshToken: refreshJwtToken,
	}, nil
}

func (jm *JwtManager) getJwtToken(tokenType string, tokenLifeTime time.Duration, now time.Time, issuer string, userId uuid.UUID) (*jwt.Token, error) {
	jwtToken := jwt.NewWithClaims(signingMethod, CustomClaims{
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userId.String(),
			Issuer:    issuer,
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenLifeTime)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	})
	key := []byte(jm.Config.JwtSecret)
	signedToken, err := jwtToken.SignedString(key)
	if err != nil {
		return nil, err
	}

	parsedToken, err := jm.Parse(signedToken)
	if err != nil {
		return nil, err
	}

	return parsedToken, nil
}
