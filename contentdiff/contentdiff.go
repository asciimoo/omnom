// Package contentdiff provides HTML comparison and difference detection functionality.
//
// This package compares HTML documents to identify changes between different versions
// of a bookmarked web page. It extracts and compares:
//   - Text content (rendered text from the page)
//   - Links (href and anchor text)
//   - Multimedia elements (images, videos)
//
// The comparison uses the Myers diff algorithm (via sergi/go-diff) to compute
// differences in text content. For links and multimedia, it performs set-based
// comparison to identify additions and removals.
//
// The package is used to show users what has changed on a bookmarked page between
// snapshots, making it easy to track content updates, new links, or removed sections.
//
// Example usage:
//
//	reader1 := strings.NewReader(oldHTML)
//	reader2 := strings.NewReader(newHTML)
//	diffs, err := contentdiff.DiffHTML(reader1, reader2)
//	if err != nil {
//	    return err
//	}
//
//	// Display text changes
//	for _, d := range diffs.Text {
//	    fmt.Printf("[%s] %s\n", d.Type, d.Text)
//	}
package contentdiff

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"strings"

	"golang.org/x/net/html"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// TextDiff represents a text difference.
type TextDiff struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

// TextDiffs is a list of TextDiff items
type TextDiffs []TextDiff

// LinkDiff represents a link difference.
type LinkDiff struct {
	Link Link   `json:"link"`
	Type string `json:"type"`
}

// LinkDiffs is a list of LinkDiff items
type LinkDiffs []LinkDiff

// Diffs contains all types of differences between two HTML documents.
type Diffs struct {
	Text       TextDiffs `json:"text"`
	Multimedia TextDiffs `json:"multimedia"`
	Link       LinkDiffs `json:"link"`
}

// HTMLContent represents extracted content from an HTML document.
type HTMLContent struct {
	Links      []Link   `json:"links"`
	Multimedia []string `json:"multimedia"`
	Text       string   `json:"text"`
}

// Link represents a hyperlink in HTML content.
type Link struct {
	Href string `json:"href"`
	Text string `json:"text"`
}

// DiffHTML compares two HTML documents and returns their differences.
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

// DiffText compares two text strings and returns their differences.
func DiffText(t1, t2 string) TextDiffs {
	dmp := diffmatchpatch.New()
	ds := dmp.DiffMain(t1, t2, false)
	r := make(TextDiffs, len(ds))
	for i, d := range dmp.DiffCleanupSemantic(ds) {
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

// DiffLink compares two sets of links and returns their differences.
func DiffLink(l1, l2 []Link) LinkDiffs {
	r := make(LinkDiffs, 0)
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

// DiffList compares two string lists and returns their differences.
func DiffList(l1, l2 []string) TextDiffs {
	r := make(TextDiffs, 0)
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

// ExtractHTMLContent extracts text, links, and multimedia from an HTML document.
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
				capture = true
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

func (lds LinkDiffs) String() string {
	r := make([]string, len(lds))
	for i, l := range lds {
		t := "removed"
		if l.Type == "+" {
			t = "added"
		}
		r[i] = fmt.Sprintf("%s %s (%s)", l.Link.Href, l.Link.Text, t)
	}
	return strings.Join(r, "\n")
}

func (tds TextDiffs) String() string {
	r := make([]string, 0, 8)
	for _, t := range tds {
		if t.Type == "0" {
			continue
		}
		r = append(r, fmt.Sprintf("%s %#v", t.Type, t.Text))
	}
	return strings.Join(r, "\n")
}
