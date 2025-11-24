// Package localization provides internationalization (i18n) support for Omnom.
//
// This package handles translation of user interface strings into multiple languages
// using the go-i18n library. Translation files are JSON-based and embedded in the
// application at build time.
//
// Supported features:
//   - Multiple language support with automatic fallback to English
//   - Parameterized message templates
//   - Language detection from Accept-Language headers
//   - Embedded translation files
//
// Translation files are stored in the locales/ directory with names like en.json,
// es.json, etc. Each file contains message IDs and their translated strings.
//
// The Localizer type provides thread-safe translation lookups for a specific
// language preference. The global SupportedLanguages slice contains metadata
// about all available translations.
//
// Example usage:
//
//	// Create a localizer for a user's preferred languages
//	loc := localization.NewLocalizer("es", "en")
//
//	// Translate a simple message
//	text := loc.Msg("welcome_message")
//
//	// Translate with parameters
//	text := loc.Msgf("greeting", "name", userName, "count", itemCount)
package localization

import (
	"embed"
	"encoding/json"
	"os"
	"strings"

	"github.com/asciimoo/omnom/utils"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

//go:embed locales/*.json
var fs embed.FS
var bundle *i18n.Bundle

// SupportedLanguages contains all available language translations.
var SupportedLanguages []*LangInfo
var defaultLocalizer *Localizer

// LangInfo contains information about a language.
type LangInfo struct {
	Abbr        string
	EnglishName string
	Name        string
}

// Localizer provides translation functionality.
type Localizer struct {
	l *i18n.Localizer
}

// NewLocalizer creates a new localizer for the specified languages.
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

// Msg translates a string
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

// Msgf translates a string with optional translation settings.
// See https://pkg.go.dev/github.com/nicksnyder/go-i18n/v2@v2.6.0/i18n#LocalizeConfig for details.
func (l *Localizer) Msgf(msg string, values ...any) string {
	data, err := utils.KVData(values...)
	if err != nil {
		log.Debug().Err(err).Str("id", msg).Msg("Invalid translation data")
		return msg
	}
	m := &i18n.LocalizeConfig{
		MessageID:    msg,
		TemplateData: data,
	}
	tm, err := l.l.Localize(m)
	if err != nil {
		tem, err := defaultLocalizer.l.Localize(m)
		if err != nil {
			log.Debug().Err(err).Str("id", msg).Msg("Missing or invalid translation")
			return msg
		}
		return tem
	}
	return tm
}
