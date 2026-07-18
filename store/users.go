package store

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type UserStore struct {
	db *sqlx.DB
}

type User struct {
	Id                   uuid.UUID `db:"id"`
	Email                string    `db:"email"`
	HashedPasswordBase64 string    `db:"hashed_password"`
	CreatedAt            time.Time `db:"created_at"`
}

func (user *User) ComparePassword(password string) error {
	hashedPassword, err := base64.StdEncoding.DecodeString(user.HashedPasswordBase64)
	if err != nil {
		return err
	}
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		return fmt.Errorf("Invalid password")
	}
	return nil
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

func (s *UserStore) CreateUser(ctx context.Context, email, password string) (*User, error) {
	const dml = "insert into users(email, hashed_password) values ($1, $2)"

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("Could not hash password: %w", err)
	}
	hashedPasswordBase64 := base64.StdEncoding.EncodeToString(bytes)

	var user User
	if err := s.db.GetContext(ctx, &user, dml, email, hashedPasswordBase64); err != nil {
		return nil, fmt.Errorf("Could not insert user: %w", err)
	}
	return &user, nil
}

func (s *UserStore) ByEmail(ctx context.Context, email string) (*User, error) {
	const sql = "select * from users where email = $1"
	var user User
	if err := s.db.GetContext(ctx, &user, sql, email); err != nil {
		return nil, fmt.Errorf("Could not find user by email %s: %w", email, err)
	}
	return &user, nil
}

func (s *UserStore) ById(ctx context.Context, userId uuid.UUID) (*User, error) {
	const sql = "select * from users where id = $1"
	var user User
	if err := s.db.GetContext(ctx, &user, sql, userId); err != nil {
		return nil, fmt.Errorf("Could not find user by id %s: %w", userId, err)
	}
	return &user, nil
}
