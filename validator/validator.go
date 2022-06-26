// SPDX-FileCopyrightText: 2021-2022 Adam Tauber, <asciimoo@gmail.com> et al.
//
// SPDX-License-Identifier: AGPL-3.0-only

package validator

import (
	"bytes"
	"errors"
	"io"

	"golang.org/x/net/html"
)

func ValidateHTML(h []byte) error {
	r := bytes.NewReader(h)
	doc := html.NewTokenizer(r)
	for {
		tt := doc.Next()
		switch tt {
		case html.ErrorToken:
			err := doc.Err()
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		case html.StartTagToken:
			tn, hasAttr := doc.TagName()
			if bytes.Equal(tn, []byte("script")) {
				return errors.New("script tag found")
			}
			if hasAttr {
				for {
					aName, aVal, moreAttr := doc.TagAttr()
					if bytes.HasPrefix(aName, []byte("on")) && len(aVal) > 0 {
						return errors.New("invalid attribute " + string(aName))
					}
					if !moreAttr {
						break
					}
				}
			}
		}
	}
}
