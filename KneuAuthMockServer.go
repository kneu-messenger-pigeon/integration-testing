package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/vitorsalgado/mocha/v3"
	"github.com/vitorsalgado/mocha/v3/expect"
	"github.com/vitorsalgado/mocha/v3/reply"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type KneuAuthMockServer struct {
	mocha         *mocha.Mocha
	updates       chan TelegramUpdate
	lastUpdateId  uint32
	lastCodeIndex uint32
}

func CreateKneuAuthMockServer(t *testing.T, clientId int, clientSecret string) *KneuAuthMockServer {
	kneuAuthMockServer := &KneuAuthMockServer{
		updates: make(chan TelegramUpdate),
	}

	configure := mocha.Configure()
	configure.Addr(kneuAuthServerAddr)

	kneuAuthMockServer.mocha = mocha.New(t, configure.Build())
	kneuAuthMockServer.mocha.Start()

	return kneuAuthMockServer
}

func (mockServer *KneuAuthMockServer) Close() {
	_ = mockServer.mocha.Close()
}

func (mockServer *KneuAuthMockServer) EmulateAuthFlow(t *testing.T, authUrlString string) {
	authUrlString = strings.ReplaceAll(authUrlString, "\\", "")

	authUrl, err := url.Parse(authUrlString)

	if err != nil || authUrl == nil {
		t.Errorf("parse auth authUrl error: %v\n", err)
		return
	}

	if !strings.HasPrefix(authUrlString, config.kneuBaseUrl) {
		t.Errorf("auth authUrl does not start with kneuBaseUrl: " + authUrlString)
		return
	}

	if authUrl.Query().Get("response_type") != "code" {
		t.Errorf("response_type is not code in auth authUrl: " + authUrlString)
		return
	}

	origRedirectUriString := authUrl.Query().Get("redirect_uri")
	state := authUrl.Query().Get("state")

	if origRedirectUriString == "" {
		t.Errorf("redirect_uri is empty in auth authUrl: " + authUrlString)
		return
	}

	if state == "" {
		t.Errorf("state is empty in auth authUrl: " + authUrlString)
		return
	}

	//	fmt.Println("redirect_uri: ", origRedirectUriString)

	redirectUri, err := authUrl.Parse(origRedirectUriString)
	if err != nil || redirectUri == nil {
		t.Errorf("parse redirectUri error: %v\n", err)
		return
	}

	atomic.AddUint32(&mockServer.lastCodeIndex, 1)
	code := "testCode" + strconv.Itoa(int(mockServer.lastCodeIndex))
	accessToken := "testAccessToken" + strconv.Itoa(int(mockServer.lastCodeIndex))
	fmt.Println(accessToken)
	redirectUriQuery := redirectUri.Query()
	redirectUriQuery.Set("code", code)
	redirectUriQuery.Set("state", state)
	redirectUri.RawQuery = redirectUriQuery.Encode()

	authorizerClientRequestsScope := mockServer.mocha.AddMocks(
		mocha.Post(expect.URLPath("/oauth/token")).
			Repeat(1).
			Header("Content-Type", expect.ToEqual("application/x-www-form-urlencoded")).
			FormField("client_id", expect.ToEqual(strconv.Itoa(config.kneuClientId))).
			FormField("client_secret", expect.ToEqual(config.kneuClientSecret)).
			FormField("grant_type", expect.ToEqual("authorization_code")).
			FormField("code", expect.ToEqual(code)).
			FormField("redirect_uri", expect.ToEqual(origRedirectUriString)).
			Reply(
				reply.OK().BodyJSON(map[string]interface{}{
					"access_token": accessToken,
					"token_type":   "Bearer",
					"expires_in":   7200,
					"user_id":      999,
				}),
			),

		mocha.Get(expect.URLPath("/api/user/me")).
			Repeat(1).
			Header("Authorization", expect.ToEqual("Bearer "+accessToken)).
			Reply(
				reply.OK().BodyJSON(map[string]interface{}{
					"id":          999,
					"group_id":    50,
					"last_name":   "Петренко",
					"first_name":  "Петр",
					"middle_name": "Петрович",
					"name":        "Петренко Петр Петрович",
					"email":       "test@example.com",
					"type":        "student",
					"gender":      "male",
					"student_id":  123,
				}),
			),
	)
	defer authorizerClientRequestsScope.Clean()

	fmt.Println("emulate success auth redirect to:", redirectUri.String())

	response, err := http.Get(redirectUri.String())
	if err != nil {
		t.Errorf("navigate to redirectUri error: %v\n", err)
		return
	}

	authorizerClientRequestsScope.AssertCalled(t)

	assert.Equal(t, 200, response.StatusCode, "response status code is not 200")

	time.Sleep(time.Second * 3)
}
