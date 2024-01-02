package tests

import (
	"context"
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
	assert.Error(t, err)
	assert.Empty(t, respRegister.GetId())
	assert.ErrorContains(t, err, "user already exists")
}

func TestLogin_invalidCredentials(t *testing.T) {
	ctx, st := suite.NewSuite(t)

	email := gofakeit.Email()
	pass := randomFakePassport()

	respRegister, err := register(ctx, st, email, pass)
	require.NoError(t, err)
	assert.NotEmpty(t, respRegister.GetId())

	respLogin, err := login(ctx, st, gofakeit.Email(), pass)
	assert.Error(t, err, "email changed")
	assert.Empty(t, respLogin.GetToken())
	assert.ErrorContains(t, err, "user not found")

	respLogin, err = login(ctx, st, email, randomFakePassport())
	assert.Error(t, err, "pass changed")
	assert.Empty(t, respLogin.GetToken())
	assert.ErrorContains(t, err, "user not found")
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
		AppId:    4,
	})
	assert.Error(t, err)
	assert.Empty(t, respLogin.GetToken())
	assert.ErrorContains(t, err, "app not found")
}

func TestRegister_validationErrors(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		password    string
		expectedErr string
	}{
		{
			name:        "Register with empty email",
			email:       "",
			password:    randomFakePassport(),
			expectedErr: "empty email",
		},
		{
			name:        "Register with empty password",
			email:       gofakeit.Email(),
			password:    "",
			expectedErr: "empty password",
		},
		{
			name:        "Register with both empty",
			email:       "",
			password:    "",
			expectedErr: "empty email",
		},
	}

	ctx, st := suite.NewSuite(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := st.AuthClient.Register(ctx, &domofon_v1.RegisterRequest{
				Email:    tt.email,
				Password: tt.password,
			})
			assert.Error(t, err)
			assert.ErrorContains(t, err, tt.expectedErr)
		})
	}
}

func TestLogin_validationErrors(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		password    string
		appId       int32
		expectedErr string
	}{
		{
			name:        "Login with empty email",
			email:       "",
			password:    randomFakePassport(),
			appId:       AppId,
			expectedErr: "empty email",
		},
		{
			name:        "Login with empty password",
			email:       gofakeit.Email(),
			password:    "",
			appId:       AppId,
			expectedErr: "empty password",
		},
		{
			name:        "Login with empty app_id",
			email:       gofakeit.Email(),
			password:    gofakeit.Email(),
			appId:       emptyAppId,
			expectedErr: "empty app_id",
		},
		{
			name:        "Login with empty fields",
			expectedErr: "empty email",
		},
	}

	ctx, st := suite.NewSuite(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := st.AuthClient.Login(ctx, &domofon_v1.LoginRequest{
				Email:    tt.email,
				Password: tt.password,
				AppId:    tt.appId,
			})
			assert.Error(t, err)
			assert.ErrorContains(t, err, tt.expectedErr)
		})
	}
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
