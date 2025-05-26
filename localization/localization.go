package localization

import (
	"embed"
	"encoding/json"
	"os"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

//go:embed locales/*.json
var fs embed.FS
var bundle *i18n.Bundle
var SupportedLanguages []*LangInfo

type LangInfo struct {
	Abbr        string
	EnglishName string
	Name        string
}

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
	d, err := fs.ReadDir("locales")
	if err != nil {
		panic(err)
	}
	SupportedLanguages = make([]*LangInfo, len(d))
	for i, f := range d {
		m, err := bundle.LoadMessageFileFS(fs, "locales/"+f.Name())
		if err != nil {
			log.Error().Err(err).Str("file", f.Name()).Msg("Failed to load translation file")
			panic(err)
		}
		if len(m.Messages) == 0 {
			continue
		}
		lang := strings.TrimSuffix(f.Name(), ".json")
		t, err := language.Parse(lang)
		if err != nil {
			log.Error().Err(err).Str("lang", lang).Msg("Failed to identify translation language")
			os.Exit(1)
		}
		SupportedLanguages[i] = &LangInfo{
			Abbr:        lang,
			EnglishName: display.English.Tags().Name(t),
			Name:        display.Self.Name(t),
		}
		m.Tag = t
	}
}

func (l *Localizer) Msg(msg string) string {
	tm, err := l.l.LocalizeMessage(&i18n.Message{
		ID: msg,
	})
	if err != nil {
		log.Debug().Err(err).Msg("Missing translation")
		return msg
	}
	return tm
}

func (l *Localizer) Data(msg string, data map[string]interface{}) string {
	tm, err := l.l.Localize(&i18n.LocalizeConfig{
		MessageID:    msg,
		TemplateData: data,
	})
	if err != nil {
		log.Debug().Err(err).Msg("Missing translation")
		return msg
	}
	return tm
}
