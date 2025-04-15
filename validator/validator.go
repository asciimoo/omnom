// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package validator

import (
	"bytes"
	"errors"
	"io"

	"golang.org/x/net/html"
)

type Result struct {
	Error        error
	HasShadowDOM bool
}

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
