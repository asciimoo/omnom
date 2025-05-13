package webapp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/asciimoo/omnom/config"
	"github.com/asciimoo/omnom/model"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

var testCfg = &config.Config{
	App: config.App{
		LogLevel:    "debug",
		TemplateDir: "../templates/",
	},
	DB: config.DB{
		Type:       "sqlite",
		Connection: ":memory:",
	},
	Server: config.Server{
		BaseURL: "https://test.com/",
	},
	ActivityPub: &config.ActivityPub{},
}

func initTestApp() *gin.Engine {
	_, _ = testCfg.ActivityPub.ExportPrivKey()
	err := model.Init(testCfg)
	if err != nil {
		panic(err)
	}
	return createEngine(testCfg)
}

func TestAPIdentity(t *testing.T) {
	router := initTestApp()
	err := model.CreateUser("test", "test@test.com")
	if !assert.Nil(t, err) {
		log.Debug().Msg("failed to create test user")
		return
	}
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", URLFor("User", "test"), nil)
	req.Header.Add("Accept", "application/activity+json")
	router.ServeHTTP(w, req)

	if !assert.Equal(t, 200, w.Code) {
		return
	}
	var ru apIdentity

	err = json.Unmarshal(w.Body.Bytes(), &ru)
	if !assert.Nil(t, err) {
		log.Debug().Bytes("body", w.Body.Bytes()).Msg("failed to parse JSON")
		return
	}
	if !assert.Equal(t, "https://test.com/users/test", ru.ID) {
		return
	}
	if !assert.Equal(t, "https://test.com/inbox/test", ru.Inbox) {
		return
	}
	if !assert.Equal(t, "https://test.com/outbox/test", ru.Outbox) {
		return
	}
}

func TestAPWebfinger(t *testing.T) {
	router := initTestApp()
	err := model.CreateUser("test", "test@test.com")
	if !assert.Nil(t, err) {
		log.Debug().Msg("failed to create test user")
		return
	}
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/.well-known/webfinger?resource=acct:test@test.com", nil)
	router.ServeHTTP(w, req)
	if !assert.Equal(t, 200, w.Code) {
		return
	}

	var wf apWebfinger
	err = json.Unmarshal(w.Body.Bytes(), &wf)
	if !assert.Nil(t, err) {
		log.Debug().Bytes("body", w.Body.Bytes()).Msg("failed to parse JSON")
		return
	}
	if !assert.Len(t, wf.Links, 2, "Expected 2 webfinger links") {
		return
	}
}
