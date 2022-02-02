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
			if err == io.EOF {
				return nil
			}
			return err
		case html.StartTagToken:
			tn, hasAttr := doc.TagName()
			if bytes.Equal(tn, []byte("script")) {
				return errors.New("Script tag found")
			}
			if hasAttr {
				for {
					aName, _, moreAttr := doc.TagAttr()
					if bytes.HasPrefix(aName, []byte("on")) {
						return errors.New("Invalid attribute " + string(aName))
					}
					if !moreAttr {
						break
					}
				}
			}
		}
	}
	return nil
}
