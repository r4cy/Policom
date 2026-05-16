package aaa

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"yadro.com/course/api/core"
)

type Auth struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func (a *AAA) Register(ctx context.Context, username string, email string, password string, role core.UserRole) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		a.log.Error("bad password")
		return core.ErrInvalidPassword
	}

	if err = a.db.AddUser(ctx, username, email, string(hash), role); err != nil {
		a.log.Error("can't to add user", "error", err)
		return err
	}
	return nil
}

func (a *AAA) Login(ctx context.Context, username string, password string) (string, error) {
	user, err := a.db.CheckUser(ctx, username)
	if err != nil {
		a.log.Error("unknown user", "username", username)
		return "", core.ErrInvalidCredentials
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password_hash), []byte(password)); err != nil {
		a.log.Error("incorrect password")
		return "", core.ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		Auth{
			UserID:   user.ID,
			Username: username,
			Email:    user.Email,
			Role:     string(user.Role),
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.tokenTTL)),
			},
		},
	)

	return token.SignedString(a.secretKey)
}

func (a *AAA) Verify(ctx context.Context, tokenString string) (core.User, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Auth{},
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				a.log.Error("unknown signing method")
				return nil, core.ErrSigningMethod
			}
			return a.secretKey, nil
		},
	)

	if err != nil || !token.Valid {
		a.log.Error("failed to parse the token", "err", err)
		return core.User{}, core.ErrInvalidToken
	}

	c := token.Claims.(*Auth)

	return core.User{
		ID:       c.UserID,
		Username: c.Username,
		Email:    c.Email,
		Role:     core.UserRole(c.Role),
	}, nil
}
