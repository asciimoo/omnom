package localization

import (
	"embed"
	"encoding/json"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*.json
var FS embed.FS
var bundle *i18n.Bundle

type Localizer struct {
	l *i18n.Localizer
}

func NewLocalizer(langs ...string) *Localizer {
	return &Localizer{
		l: i18n.NewLocalizer(bundle, langs...),
	}
}

func init() {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	d, err := FS.ReadDir("locales")
	if err != nil {
		panic(err)
	}
	for _, f := range d {
		m, err := bundle.LoadMessageFileFS(FS, "locales/"+f.Name())
		if err != nil {
			panic(err)
		}
		lang := strings.TrimSuffix(f.Name(), ".json")
		t, err := language.Parse(lang)
		if err != nil {
			panic(err)
		}
		m.Tag = t
	}
}

func (l *Localizer) Str(msg string) string {
	tm, err := l.l.LocalizeMessage(&i18n.Message{
		ID: msg,
	})
	if err != nil {
		return msg
	}
	return tm
}
