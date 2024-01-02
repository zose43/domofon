package tests

import (
	"context"
	"domofon/internal/services/auth"
	"domofon/tests/suite"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	domofon_v1 "github.com/zose43/domofon-proto/out/go"
	"testing"
	"time"
)

const (
	emptyAppId     = 0
	AppId          = 1
	appSecret      = "test-secret"
	passDefaultLen = 10
)

// todo fx migrations for tests
func TestRegisterLogin_Login_happyPath(t *testing.T) {
	ctx, st := suite.NewSuite(t)

	email := gofakeit.Email()
	pass := randomFakePassport()

	respRegister, err := register(ctx, st, email, pass)
	require.NoError(t, err)
	assert.NotEmpty(t, respRegister.GetId())

	respLogin, err := login(ctx, st, email, pass)
	loginTime := time.Now()

	require.NoError(t, err)
	token := respLogin.GetToken()
	assert.NotEmpty(t, token)

	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	assert.True(t, ok)

	assert.Equal(t, respRegister.GetId(), int64(claims["uid"].(float64)))
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, AppId, int(claims["app"].(float64)))

	const deltaSec = 1
	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), int64(claims["exp"].(float64)), deltaSec)
}

func TestRegister_alreadyUserExist(t *testing.T) {
	ctx, st := suite.NewSuite(t)

	email := gofakeit.Email()
	pass := randomFakePassport()

	respRegister, err := register(ctx, st, email, pass)
	require.NoError(t, err)
	assert.NotEmpty(t, respRegister.GetId())

	respRegister, err = register(ctx, st, email, pass)
	assert.Error(t, auth.ErrUserExists)
	assert.Empty(t, respRegister.GetId())
}

func TestLogin_invalidCredentials(t *testing.T) {
	ctx, st := suite.NewSuite(t)

	email := gofakeit.Email()
	pass := randomFakePassport()

	respRegister, err := register(ctx, st, email, pass)
	require.NoError(t, err)
	assert.NotEmpty(t, respRegister.GetId())

	respLogin, err := login(ctx, st, gofakeit.Email(), pass)
	assert.Error(t, auth.ErrInvalidCredentials, "email changed")
	assert.Empty(t, respLogin.GetToken())

	respLogin, err = login(ctx, st, email, randomFakePassport())
	assert.Error(t, auth.ErrInvalidCredentials, "pass changed")
	assert.Empty(t, respLogin.GetToken())
}

func TestLogin_invalidApp(t *testing.T) {
	ctx, st := suite.NewSuite(t)

	email := gofakeit.Email()
	pass := randomFakePassport()

	respRegister, err := register(ctx, st, email, pass)
	require.NoError(t, err)
	assert.NotEmpty(t, respRegister.GetId())

	respLogin, err := st.AuthClient.Login(ctx, &domofon_v1.LoginRequest{
		Email:    email,
		Password: pass,
		AppId:    emptyAppId,
	})
	assert.Error(t, auth.ErrInvalidApp)
	assert.Empty(t, respLogin.GetToken())
}

func login(ctx context.Context,
	st *suite.Suite,
	email string,
	pass string) (*domofon_v1.LoginResponse, error) {
	return st.AuthClient.Login(ctx, &domofon_v1.LoginRequest{
		Email:    email,
		Password: pass,
		AppId:    AppId,
	})
}

func register(
	ctx context.Context,
	st *suite.Suite,
	email string,
	pass string,
) (*domofon_v1.RegisterResponse, error) {
	return st.AuthClient.Register(ctx, &domofon_v1.RegisterRequest{
		Email:    email,
		Password: pass,
	})
}

func randomFakePassport() string {
	return gofakeit.Password(true, true, true, false, false, passDefaultLen)
}
