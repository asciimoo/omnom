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
var defaultLocalizer *Localizer

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
	SupportedLanguages = make([]*LangInfo, 0)
	for _, f := range d {
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
		SupportedLanguages = append(SupportedLanguages, &LangInfo{
			Abbr:        lang,
			EnglishName: display.English.Tags().Name(t),
			Name:        display.Self.Name(t),
		})
		m.Tag = t
	}
	defaultLocalizer = NewLocalizer("en")
}

func (l *Localizer) Msg(msg string) string {
	m := &i18n.Message{
		ID: msg,
	}
	tm, err := l.l.LocalizeMessage(m)
	if err != nil {
		tem, err := defaultLocalizer.l.LocalizeMessage(m)
		if err != nil {
			log.Debug().Err(err).Str("id", msg).Msg("Missing translation")
			return msg
		}
		return tem
	}
	return tm
}

func (l *Localizer) Data(msg string, data map[string]interface{}) string {
	m := &i18n.LocalizeConfig{
		MessageID:    msg,
		TemplateData: data,
	}
	tm, err := l.l.Localize(m)
	if err != nil {
		tem, err := defaultLocalizer.l.Localize(m)
		if err != nil {
			log.Debug().Err(err).Str("id", msg).Msg("Missing translation")
			return msg
		}
		return tem
	}
	return tm
}
