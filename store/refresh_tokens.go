package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type RefreshTokenStore struct {
	db *sqlx.DB
}

func NewRefreshTokenStore(db *sql.DB) *RefreshTokenStore {
	return &RefreshTokenStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

type RefreshToken struct {
	UserId      uuid.UUID `db:"user_id"`
	HashedToken string    `db:"hashed_token"`
	CreatedAt   time.Time `db:"created_at"`
	ExpiresAt   time.Time `db:"expires_at"`
}

func (s *RefreshTokenStore) getBase64HashFromToken(ctx context.Context, token *jwt.Token) (string, error) {
	h := sha256.New()
	h.Write([]byte(token.Raw))
	hashedBytes := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(hashedBytes), nil
}

func (s *RefreshTokenStore) Create(ctx context.Context, userId uuid.UUID, token *jwt.Token) (*RefreshToken, error) {
	const insertSql = `insert into refresh_tokens(user_id, hashed_token, expires_at) values($1, $2, $3) RETURNING *;`

	hashedTokenBase64, err := s.getBase64HashFromToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get base64 encoded token hash: %w", err)
	}

	expiresAt, err := token.Claims.GetExpirationTime()
	if err != nil {
		return nil, fmt.Errorf("failed to extract expiration time: %w", err)
	}

	var refreshToken RefreshToken
	if err := s.db.GetContext(ctx, &refreshToken, insertSql, userId, hashedTokenBase64, expiresAt.Time); err != nil {
		return nil, fmt.Errorf("failed to create refresh token record: %w", err)
	}

	return &refreshToken, nil
}

func (s *RefreshTokenStore) ByPrimaryKey(ctx context.Context, userId uuid.UUID, token *jwt.Token) (*RefreshToken, error) {
	const query = `select * from refresh_tokens where user_id=$1 and hashed_token=$2;`

	hashedTokenBase64, err := s.getBase64HashFromToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get base64 encoded token hash: %w", err)
	}

	var refreshToken RefreshToken
	if err := s.db.GetContext(ctx, &refreshToken, query, userId, hashedTokenBase64); err != nil {
		return nil, fmt.Errorf("failed to fetch hash_token %s record for user %s: %w", hashedTokenBase64, userId, err)
	}
	return &refreshToken, nil
}

func (s *RefreshTokenStore) Delete(ctx context.Context, userId uuid.UUID) (sql.Result, error) {
	const query = `delete from refresh_tokens where user_id=$1;`
	result, err := s.db.ExecContext(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to delete refresh_token by user id %s: %w", userId, err)
	}
	return result, nil
}
