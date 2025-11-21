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
	"net/url"
	"strings"

	"github.com/asciimoo/omnom/docs"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
)

const indexDocName = "index"

type docPage struct {
	Name    string
	Title   string
	Content template.HTML
}

var docTOCNames = []string{
	"index",
	"addon",
}

var docPages []*docPage
var docTOC []*docPage
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
	err = r.rewriteAST(n)
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

func (r *tRenderer) rewriteAST(n ast.Node) error {
	if n.ChildCount() < 1 {
		return nil
	}
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch c.Kind() {
		case ast.KindHeading:
			c.SetAttributeString("class", "title")
		case ast.KindList:
			c.SetAttributeString("class", "list")
		case ast.KindLink:
			n := c.(*ast.Link)
			n.Destination = []byte(resolveDocURL(string(n.Destination)))
		}
		err := r.rewriteAST(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func resolveDocURL(u string) string {
	pu, err := url.Parse(u)
	if err != nil {
		log.Error().Err(err).Str("URL", u).Msg("Failed to parse url in documentation")
		panic(err)
	}
	if pu.Scheme != "" {
		return u
	}
	if strings.HasPrefix(pu.Path, "/") {
		msg := "Invalid doc link. Use [api_endpoint_name]/[params] format for in-site doc references"
		log.Error().Err(err).Str("URL", u).Msg(msg)
		panic(msg)
	}
	if strings.Contains(pu.Path, "/") {
		parts := strings.Split(pu.Path, "/")
		u = URLFor(parts[0], parts[1:]...)
	} else {
		u = URLFor(pu.Path)
	}
	if pu.RawQuery != "" {
		u += "?" + pu.RawQuery
	}
	return u
}

func initDocs() {
	pages, err := docs.FS.ReadDir(".")
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse documentation pages")
		panic(err)
	}
	docPages = make([]*docPage, len(pages))
	// parse doc pages
	for i, p := range pages {
		r := &tRenderer{r: goldmark.DefaultRenderer()}
		md := goldmark.New(goldmark.WithRenderer(r))
		if !strings.HasSuffix(p.Name(), ".md") {
			panic("Invalid filename: " + p.Name())
		}
		name := strings.TrimSuffix(p.Name(), ".md")
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
			Content: template.HTML(buf.String()), //nolint: gosec //trusted source
		}
		if name == indexDocName {
			indexDocPage = dp
		}
		docPages[i] = dp
	}
	if indexDocPage == nil {
		panic("Missing index documentation page")
	}
	// prepare TOC
	docTOC = make([]*docPage, len(docTOCNames))
	for i, n := range docTOCNames {
		found := false
		for _, d := range docPages {
			if d.Name == n {
				docTOC[i] = d
				found = true
				break
			}
		}
		if !found {
			panic("No document name found with name " + n)
		}
	}
}

func documentation(c *gin.Context) {
	vars := gin.H{
		"Doc":   indexDocPage,
		"Pages": docPages,
		"Index": indexDocPage,
		"TOC":   docTOC,
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
