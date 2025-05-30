package webapp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/asciimoo/omnom/config"
	"github.com/asciimoo/omnom/model"
	"github.com/asciimoo/omnom/storage"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

var testCfg = &config.Config{
	App: config.App{
		LogLevel: "debug",
	},
	DB: config.DB{
		Type:       "sqlite",
		Connection: ":memory:",
	},
	Storage: config.Storage{
		Type: "fs",
	},
	Server: config.Server{
		BaseURL: "https://test.com/",
	},
	ActivityPub: &config.ActivityPub{},
}

var testActorJSON = []byte(`{
    "@context": ["https://www.w3.org/ns/activitystreams", { "@language": "en- GB" }],
    "type": "Person",
    "id": "https://test.com/testuser",
    "outbox": "https://test.com/testuser/outbox",
    "following": "https://test.com/testuser/following",
    "followers": "https://test.com/testuser/followers",
    "inbox": "https://test.com/testuser/inbox",
    "preferredUsername": "testuser",
    "name": "testuser",
    "summary": "testuser summary",
    "icon": {
      "url": "https://test.com/testuser/images/me.png"
    },
    "publicKey": {
      "@context": "https://w3id.org/security/v1",
      "@type": "Key",
      "id": "https://test.com/testuser#main-key",
      "owner": "https://test.com/testuser",
      "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAsniRFKHTS2zn3mpe/Ic6\nqoT42Yz+sqjIESU1yIKFQgcsOo0w8eqoiZK6r+oWGc28ZCQ1KHaz123Z7bTazxSP\n0JEtLIpePZLWkcWc0Ryx/ZACQ6c7XtKi5Wq/zIhT0+XJGzkmbYsSiKOMDodY7T98\nbKC/R4lTeQWv4aiAYccTr6KwIGAijz9aOYbSD69h80HwYiru2RE8bcPol7ZLN55R\nQ1dwY/i7QSTwaFFUoFKBc1tFne9Lktgr6mA0WxJ4hEkTF/N5leDc2q/IorOY6STt\nnieo1QIcKnf4w+I4LjutK7l8I38Sn+6YHVF64B/lXyMsBenZ2r56i94mlEvdsCtd\n+QIDAQAB\n-----END PUBLIC KEY-----\n"
    }
}`)

func initTestApp() *gin.Engine {
	_, _ = testCfg.ActivityPub.ExportPrivKey()
	err := model.Init(testCfg)
	if err != nil {
		panic(err)
	}
	err = storage.Init(testCfg.Storage)
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

func TestAPSigHeaderParse(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Set("config", testCfg)
	ctx.Request = &http.Request{
		Header: make(http.Header),
		URL: &url.URL{
			Path: "/test",
		},
	}
	testHeader := `keyId="https://my.example.com/actor#main-key",headers="(request-target) host date",signature="aabbccdd"`
	ctx.Request.Header.Set("Signature", testHeader)
	ctx.Request.Header.Set("Digest", "testdigest")
	sig, _, err := apParseSigHeader(ctx, "testdigest")
	if !assert.Nil(t, err) {
		log.Debug().Msg("failed to parse signature header")
		return
	}
	if !assert.Equal(t, sig, "aabbccdd") {
		log.Debug().Msg("failed to parse signature")
		return
	}
}

func TestAPActorOutbox(t *testing.T) {
	router := initTestApp()
	err := model.CreateUser("test", "test@test.com")
	if !assert.Nil(t, err) {
		log.Debug().Msg("failed to create test user")
		return
	}
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", URLFor("activitypub outbox", "test"), nil)
	req.Header.Add("Accept", "application/activity+json")
	router.ServeHTTP(w, req)
	var o apOutbox

	err = json.Unmarshal(w.Body.Bytes(), &o)
	if !assert.Nil(t, err) {
		log.Debug().Bytes("body", w.Body.Bytes()).Msg("failed to parse JSON")
		return
	}
	if !assert.Equal(t, o.ID, "https://test.com/users/test") {
		log.Debug().Msg("failed to get outbox ID")
		return
	}
}

func TestAPActorParse(t *testing.T) {
	i := &apIdentity{}
	err := json.Unmarshal(testActorJSON, i)
	if !assert.Nil(t, err) {
		log.Debug().Msg("failed to parse actor")
		return
	}
	if !assert.Equal(t, i.Name, "testuser") {
		log.Debug().Msg("failed to get actor's name")
		return
	}
	if !assert.Equal(t, i.Inbox, "https://test.com/testuser/inbox") {
		log.Debug().Msg("failed to get actor's inbox")
		return
	}
}

//func TestAPActorFetch(t *testing.T) {
//	cfg, err := config.Load("../config.yml")
//	if !assert.Nil(t, err) {
//		log.Debug().Msg("failed to load config")
//		return
//	}
//	privBytes, err := os.ReadFile("../" + cfg.ActivityPub.PrivKeyPath)
//	if !assert.Nil(t, err) {
//		log.Debug().Msg("failed to read privk")
//		return
//	}
//	err = cfg.ActivityPub.ParsePrivKey(privBytes)
//	k := cfg.ActivityPub.PrivK
//	a, err := apFetchActor("https://merveilles.town/users/bouncepaw", "https://omnom.zone/users/testuser#key", k)
//	if !assert.Nil(t, err) {
//		log.Debug().Msg("failed to fetch actor")
//		return
//	}
//	fmt.Println(a.ID, a.Inbox)
//}
