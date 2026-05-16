package aaa

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"yadro.com/course/api/core"
)

func TestMain(t *testing.T) {
	t.Run("ping test", TestPing)
	t.Run("good Auth token test", TestVerifuValidToken)
	t.Run("bad Auth token test", TestVerifuInvalidToken)

	t.Run("expired token Auth test", TestVerifuExpiredToken)

}

func makeAAAClient() *AAA {
	return &AAA{
		secretKey: []byte("gopher"),
		tokenTTL:  time.Hour,
		log:       slog.Default(),
	}
}

func makeToken(t *testing.T, secret []byte, expiresAt time.Time) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Auth{
		UserID:   "user-1",
		Username: "roman",
		Email:    "roman@example.com",
		Role:     string(core.RoleAdmin),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	})

	tokenString, err := token.SignedString(secret)
	require.NoError(t, err)

	return tokenString
}

func TestPing(t *testing.T) {
	aaa := makeAAAClient()
	err := aaa.Ping(context.Background())
	require.NoError(t, err, "should be nil")
}

func TestVerifuValidToken(t *testing.T) {
	a := makeAAAClient()
	token := makeToken(t, a.secretKey, time.Now().Add(time.Hour))

	user, err := a.Verify(context.Background(), token)

	require.NoError(t, err)
	require.Equal(t, "user-1", user.ID)
	require.Equal(t, "roman", user.Username)
	require.Equal(t, "roman@example.com", user.Email)
	require.Equal(t, core.RoleAdmin, user.Role)
}

func TestVerifuInvalidToken(t *testing.T) {
	a := makeAAAClient()

	user, err := a.Verify(context.Background(), "not-a-token")

	require.ErrorIs(t, err, core.ErrInvalidToken)
	require.Empty(t, user)
}

func TestVerifuExpiredToken(t *testing.T) {
	a := makeAAAClient()
	token := makeToken(t, a.secretKey, time.Now().Add(-time.Hour))

	user, err := a.Verify(context.Background(), token)

	require.ErrorIs(t, err, core.ErrInvalidToken)
	require.Empty(t, user)
}
