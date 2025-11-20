// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package webapp

import (
	"bytes"
	"errors"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/asciimoo/omnom/docs"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
)

type docPage struct {
	Name    string
	Title   string
	Content template.HTML
}

var docPages []*docPage
var indexDocPage *docPage

type tRenderer struct {
	r     renderer.Renderer
	Title string
}

func (r *tRenderer) Render(w io.Writer, source []byte, n ast.Node) error {
	err := r.extractTitle(source, n)
	if err != nil {
		return err
	}
	err = r.addClasses(n)
	if err != nil {
		return err
	}
	return r.r.Render(w, source, n)
}

func (r *tRenderer) AddOptions(os ...renderer.Option) {
	r.r.AddOptions(os...)
}

func (r *tRenderer) extractTitle(source []byte, n ast.Node) error {
	if n.ChildCount() < 1 {
		return errors.New("missing children")
	}
	if n.FirstChild().Kind() != ast.KindHeading {
		return errors.New("first element must be a header")
	}
	h := n.FirstChild()
	if h.ChildCount() < 1 {
		return errors.New("missing header children")
	}
	for c := h.FirstChild(); c != nil; c = c.NextSibling() {
		if c.Kind() == ast.KindText {
			n := c.(*ast.Text)
			r.Title += string(n.Segment.Value(source))
		}
	}
	if r.Title == "" {
		return errors.New("missing title text")
	}
	return nil
}

func (r *tRenderer) addClasses(n ast.Node) error {
	if n.ChildCount() < 1 {
		return nil
	}
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch c.Kind() {
		case ast.KindHeading:
			c.SetAttributeString("class", "title")
		case ast.KindList:
			c.SetAttributeString("class", "list")
		}
		err := r.addClasses(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	pages, err := docs.FS.ReadDir(".")
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse documentation pages")
		panic(err)
	}
	docPages = make([]*docPage, len(pages)-1)
	for i, p := range pages {
		r := &tRenderer{r: goldmark.DefaultRenderer()}
		md := goldmark.New(goldmark.WithRenderer(r))
		if !strings.HasSuffix(p.Name(), ".md") {
			panic("Invalid filename: " + p.Name())
		}
		name := strings.TrimSuffix(p.Name(), ".md")
		if name != "index" {
			// TODO more accurate validation
			if len(name) < 5 {
				msg := "Page filename should start with 'XXX-' prefix where X is a digit"
				log.Error().Str("page", p.Name()).Msg(msg)
				panic(msg)
			}
			name = name[4:]
		}
		var buf bytes.Buffer
		fc, err := docs.FS.ReadFile(p.Name())
		if err != nil {
			log.Error().Err(err).Str("page", p.Name()).Msg("Failed to read documentation page")
			panic(err)
		}
		err = md.Convert(fc, &buf)
		if err != nil {
			log.Error().Err(err).Str("page", p.Name()).Msg("Failed to parse documentation page")
			panic(err)
		}
		dp := &docPage{
			Name:    name,
			Title:   r.Title,
			Content: template.HTML(buf.String()),
		}
		if name == "index" {
			indexDocPage = dp
		} else {
			docPages[i] = dp
		}
	}
	if indexDocPage == nil {
		panic("Missing index documentation page")
	}
}

func documentation(c *gin.Context) {
	vars := gin.H{
		"Doc":   indexDocPage,
		"Pages": docPages,
		"Index": indexDocPage,
	}
	page := strings.TrimPrefix(c.Param("page"), "/")
	log.Debug().Msg(page)
	for _, p := range docPages {
		if p.Name == page {
			vars["Doc"] = p
		}
	}
	render(c, http.StatusOK, "docs", vars)
}
