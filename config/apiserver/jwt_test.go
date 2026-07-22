package apiserver_test

import (
	"go_sqs_pqsql_s3_project/config"
	"go_sqs_pqsql_s3_project/config/apiserver"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestJwtManager(t *testing.T) {
	conf, err := config.New()
	require.NoError(t, err)

	jwtManager := apiserver.NewJwtManger(conf)
	userId := uuid.New()
	tokenPair, err := jwtManager.GenerateTokenPair(userId)
	require.NoError(t, err)

	// access
	require.True(t, jwtManager.IsAccessToken(tokenPair.AccessToken))

	accessTokenSubject, err := tokenPair.AccessToken.Claims.GetSubject()
	require.NoError(t, err)
	require.Equal(t, userId.String(), accessTokenSubject)

	accessTokenIssuer, err := tokenPair.AccessToken.Claims.GetIssuer()
	require.NoError(t, err)
	require.Equal(t, "http://"+conf.ApiServerHost+":"+conf.ApiServerPort, accessTokenIssuer)

	parsedAccessToken, err := jwtManager.Parse(tokenPair.AccessToken.Raw)
	require.NoError(t, err)
	require.Equal(t, tokenPair.AccessToken, parsedAccessToken)

	// refresh
	require.False(t, jwtManager.IsAccessToken(tokenPair.RefreshToken))

	refreshTokenSubject, err := tokenPair.RefreshToken.Claims.GetSubject()
	require.NoError(t, err)
	require.Equal(t, userId.String(), refreshTokenSubject)

	refreshTokenIssuer, err := tokenPair.RefreshToken.Claims.GetIssuer()
	require.NoError(t, err)
	require.Equal(t, "http://"+conf.ApiServerHost+":"+conf.ApiServerPort, refreshTokenIssuer)

	parsedRefreshToken, err := jwtManager.Parse(tokenPair.RefreshToken.Raw)
	require.NoError(t, err)
	require.Equal(t, tokenPair.RefreshToken, parsedRefreshToken)

}
