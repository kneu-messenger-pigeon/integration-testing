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
)

type KneuAuthMockServer struct {
	mocha         *mocha.Mocha
	updates       chan TelegramUpdate
	lastUpdateId  uint32
	lastCodeIndex uint32
}

type FakeUser struct {
	Id         int
	StudentId  int
	GroupId    int
	LastName   string
	FirstName  string
	MiddleName string
	Gender     string
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

func (mockServer *KneuAuthMockServer) EmulateAuthFlow(t *testing.T, authUrlString string, fakeUser *FakeUser) {
	if authUrlString == "" {
		t.Errorf("authUrlString is empty")
		t.FailNow()
		return
	}

	authUrlString = strings.ReplaceAll(authUrlString, "\\", "")

	authUrl, err := url.Parse(authUrlString)

	if err != nil || authUrl == nil {
		t.Errorf("parse auth authUrl error: %v\n", err)
		t.FailNow()
		return
	}

	if !strings.HasPrefix(authUrlString, config.kneuBaseUrl) {
		t.Errorf("auth authUrl does not start with kneuBaseUrl: " + authUrlString)
		t.FailNow()
		return
	}

	if authUrl.Query().Get("response_type") != "code" {
		t.Errorf("response_type is not code in auth authUrl: " + authUrlString)
		t.FailNow()
		return
	}

	origRedirectUriString := authUrl.Query().Get("redirect_uri")
	state := authUrl.Query().Get("state")

	if origRedirectUriString == "" {
		t.Errorf("redirect_uri is empty in auth authUrl: " + authUrlString)
		t.FailNow()
		return
	}

	if state == "" {
		t.Errorf("state is empty in auth authUrl: " + authUrlString)
		t.FailNow()
		return
	}

	//	fmt.Println("redirect_uri: ", origRedirectUriString)

	redirectUri, err := authUrl.Parse(origRedirectUriString)
	if err != nil || redirectUri == nil {
		t.Errorf("parse redirectUri error: %v\n", err)
		t.FailNow()
		return
	}

	atomic.AddUint32(&mockServer.lastCodeIndex, 1)
	code := "testCode" + strconv.Itoa(int(mockServer.lastCodeIndex))
	accessToken := "testAccessToken" + strconv.Itoa(int(mockServer.lastCodeIndex))

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
					"user_id":      fakeUser.Id,
				}),
			),

		mocha.Get(expect.URLPath("/api/user/me")).
			Repeat(1).
			Header("Authorization", expect.ToEqual("Bearer "+accessToken)).
			Reply(
				reply.OK().BodyJSON(map[string]interface{}{
					"id":          fakeUser.Id,
					"group_id":    fakeUser.GroupId,
					"student_id":  fakeUser.StudentId,
					"last_name":   fakeUser.LastName,
					"first_name":  fakeUser.FirstName,
					"middle_name": fakeUser.MiddleName,
					"name":        fakeUser.LastName + " " + fakeUser.FirstName + " " + fakeUser.MiddleName,
					"gender":      fakeUser.Gender,
					"email":       "user" + strconv.Itoa(fakeUser.Id) + "@example.com",
					"type":        "student",
				}),
			),
	)
	defer authorizerClientRequestsScope.Clean()

	fmt.Println("emulate success auth redirect to:", redirectUri.String())

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if req.URL.Hostname() != redirectUri.Hostname() { // stop after outside redirect
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	response, err := client.Get(redirectUri.String())
	if err != nil {
		t.Errorf("navigate to redirectUri error: %v\n", err)
		return
	}

	authorizerClientRequestsScope.AssertCalled(t)

	assert.Equal(t, 302, response.StatusCode, "response status code is not 302")
	assert.Equal(t, "https://t.me/test?start", response.Header.Get("Location"), "unexpected redirect location")

	//	time.Sleep(time.Second * 5)
}
