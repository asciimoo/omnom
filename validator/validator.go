// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

// Package validator provides HTML validation and security checking for user content.
//
// This package validates HTML content to ensure it's safe to store and display.
// It checks for:
//   - Script tags that could execute arbitrary JavaScript
//   - Event handler attributes (onclick, onload, etc.)
//   - Shadow DOM usage (marked by omnomshadowroot attribute)
//
// The validator is used when importing bookmarks or snapshots to prevent XSS
// attacks and other security issues. It parses HTML using golang.org/x/net/html
// and reports any security concerns.
//
// Example usage:
//
//	result := validator.ValidateHTML(htmlContent)
//	if result.Error != nil {
//	    log.Printf("Invalid HTML: %v", result.Error)
//	    return
//	}
//	if result.HasShadowDOM {
//	    log.Println("Content uses Shadow DOM")
//	}
package validator

import (
	"bytes"
	"errors"
	"io"

	"golang.org/x/net/html"
)

// Result contains HTML validation results.
type Result struct {
	Error        error
	HasShadowDOM bool
}

// ValidateHTML validates HTML content for security issues.
func ValidateHTML(h []byte) Result {
	ret := Result{}
	r := bytes.NewReader(h)
	doc := html.NewTokenizer(r)
	for {
		tt := doc.Next()
		switch tt {
		case html.ErrorToken:
			err := doc.Err()
			if errors.Is(err, io.EOF) {
				return ret
			}
			ret.Error = err
			return ret
		case html.StartTagToken:
			tn, hasAttr := doc.TagName()
			if bytes.Equal(tn, []byte("script")) {
				ret.Error = errors.New("script tag found")
				return ret
			}
			if hasAttr {
				for {
					aName, aVal, moreAttr := doc.TagAttr()
					if bytes.HasPrefix(aName, []byte("on")) && len(aVal) > 0 {
						ret.Error = errors.New("invalid attribute " + string(aName))
						return ret
					}
					if bytes.Equal(aName, []byte("omnomshadowroot")) {
						ret.HasShadowDOM = true
					}
					if !moreAttr {
						break
					}
				}
			}
		}
	}
	return ret
}
