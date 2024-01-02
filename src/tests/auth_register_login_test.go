package tests

import (
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

	respRegistration, err := st.AuthClient.Register(
		ctx,
		&domofon_v1.RegisterRequest{
			Email:    email,
			Password: pass,
		},
	)

	require.NoError(t, err)
	assert.NotEmpty(t, respRegistration.GetId())

	respLogin, err := st.AuthClient.Login(
		ctx,
		&domofon_v1.LoginRequest{
			Email:    email,
			Password: pass,
			AppId:    AppId,
		},
	)

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

	assert.Equal(t, respRegistration.GetId(), int64(claims["uid"].(float64)))
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, AppId, int(claims["app"].(float64)))

	const deltaSec = 1
	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), int64(claims["exp"].(float64)), deltaSec)
}

func randomFakePassport() string {
	return gofakeit.Password(true, true, true, false, false, passDefaultLen)
}
