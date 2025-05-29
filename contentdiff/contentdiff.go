package contentdiff

import (
	"bytes"
	"io"
	"slices"
	"strings"

	"golang.org/x/net/html"

	"github.com/sergi/go-diff/diffmatchpatch"
)

type TextDiff struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

type LinkDiff struct {
	Link Link   `json:"link"`
	Type string `json:"type"`
}

type Diffs struct {
	Text       []TextDiff `json:"text"`
	Multimedia []TextDiff `json:"multimedia"`
	Link       []LinkDiff `json:"link"`
}

type HTMLContent struct {
	Links      []Link   `json:"links"`
	Multimedia []string `json:"multimedia"`
	Text       string   `json:"text"`
}

type Link struct {
	Href string `json:"href"`
	Text string `json:"text"`
}

func DiffHTML(r1, r2 io.Reader) (*Diffs, error) {
	c1 := ExtractHTMLContent(r1)
	c2 := ExtractHTMLContent(r2)
	ds := &Diffs{
		Text:       DiffText(c1.Text, c2.Text),
		Multimedia: DiffList(c1.Multimedia, c2.Multimedia),
		Link:       DiffLink(c1.Links, c2.Links),
	}
	return ds, nil
}

func DiffText(t1, t2 string) []TextDiff {
	dmp := diffmatchpatch.New()
	ds := dmp.DiffMain(t1, t2, false)
	r := make([]TextDiff, len(ds))
	for i, d := range ds {
		r[i] = TextDiff{
			Text: d.Text,
		}
		switch d.Type {
		case diffmatchpatch.DiffDelete:
			r[i].Type = "-"
		case diffmatchpatch.DiffInsert:
			r[i].Type = "+"
		case diffmatchpatch.DiffEqual:
			r[i].Type = "0"
		}
	}
	return r
}

func DiffLink(l1, l2 []Link) []LinkDiff {
	r := make([]LinkDiff, 0)
	for _, v1 := range l1 {
		found := false
		for _, v2 := range l2 {
			if v1.Href == v2.Href {
				found = true
				break
			}
		}
		if found {
			continue
		}
		dup := false
		for _, v2 := range r {
			if v1.Href == v2.Link.Href {
				dup = true
				break
			}
		}
		if dup {
			continue
		}
		r = append(r, LinkDiff{Link: v1, Type: "-"})
	}
	for _, v1 := range l2 {
		found := false
		for _, v2 := range l1 {
			if v1.Href == v2.Href {
				found = true
				break
			}
		}
		if found {
			continue
		}
		dup := false
		for _, v2 := range r {
			if v1.Href == v2.Link.Href {
				dup = true
				break
			}
		}
		if dup {
			continue
		}
		r = append(r, LinkDiff{Link: v1, Type: "+"})
	}
	return r
}

func DiffList(l1, l2 []string) []TextDiff {
	r := make([]TextDiff, 0)
	for _, v := range l1 {
		if !slices.Contains(l2, v) {
			r = append(r, TextDiff{Text: v, Type: "-"})
		}
	}
	for _, v := range l2 {
		if !slices.Contains(l1, v) {
			r = append(r, TextDiff{Text: v, Type: "+"})
		}
	}
	return r
}

func ExtractHTMLContent(r io.Reader) *HTMLContent {
	doc := html.NewTokenizer(r)
	capture := false
	body := false
	end := false
	inA := false
	c := &HTMLContent{
		Multimedia: make([]string, 0, 16),
		Links:      make([]Link, 0, 16),
	}
	strs := make([]string, 0, 16)
	for {
		tt := doc.Next()
		switch tt {
		case html.ErrorToken:
			end = true
		case html.StartTagToken:
			tn, hasAttr := doc.TagName()
			switch string(tn) {
			case "body":
				body = true
			case "style", "script":
				capture = false
			case "img":
				if !hasAttr {
					break
				}
				for {
					aName, aVal, moreAttr := doc.TagAttr()
					if bytes.Equal(aName, []byte("src")) {
						c.Multimedia = append(c.Multimedia, string(aVal))
						break
					}
					if !moreAttr {
						break
					}
				}
			case "a":
				if !hasAttr {
					break
				}
				for {
					aName, aVal, moreAttr := doc.TagAttr()
					if bytes.Equal(aName, []byte("href")) {
						inA = true
						l := Link{Href: string(aVal)}
						c.Links = append(c.Links, l)
						break
					}
					if !moreAttr {
						break
					}
				}
			}
		case html.EndTagToken:
			tn, _ := doc.TagName()
			if bytes.Equal(tn, []byte("body")) {
				body = false
			}
			if bytes.Equal(tn, []byte("style")) {
				capture = true
			}
			if bytes.Equal(tn, []byte("script")) {
				capture = true
			}
			if bytes.Equal(tn, []byte("a")) {
				inA = false
			}
		case html.TextToken:
			if !capture || !body {
				break
			}
			if inA {
				c.Links[len(c.Links)-1].Text += string(doc.Text())
			} else {
				strs = append(strs, string(doc.Text()))
			}
		}
		if end {
			break
		}
	}
	c.Text = strings.Join(strs, "")
	return c
}
