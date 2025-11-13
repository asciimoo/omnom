package feed

import (
	"net/url"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestSanitize(t *testing.T) {
	base, err := url.Parse("https://omnom.zone/")
	if err != nil {
		panic(err)
	}
	if !assert.Equal(t,
		`<a href="https://omnom.zone/test.html">test</a>`,
		sanitizeHTML(base, `<a href="/test.html" onclick="alert('!');">test</a><script>alert("!");</script>`),
	) {
		log.Error().Msg("Failed to resolve partial URL")
		return
	}
}

func TestUTMRemove(t *testing.T) {
	base, err := url.Parse("https://omnom.zone/")
	if err != nil {
		panic(err)
	}
	if !assert.Equal(t,
		`https://omnom.zone/test.html?a=b`,
		resolveURL(base, "/test.html?utm_source=xy&a=b"),
	) {
		log.Error().Msg("Failed to remove UTM param")
		return
	}
}
