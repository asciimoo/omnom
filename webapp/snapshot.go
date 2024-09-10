// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package webapp

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"

	"github.com/asciimoo/omnom/model"
	"github.com/asciimoo/omnom/storage"

	"github.com/gin-gonic/gin"

	"github.com/gin-gonic/contrib/sessions"
)

var resourceAttributes map[string]string = map[string]string{
	"img":    "src",
	"link":   "href",
	"iframe": "src",
}

func snapshotWrapper(c *gin.Context) {
	sid, ok := c.GetQuery("sid")
	if !ok {
		return
	}
	bid, ok := c.GetQuery("bid")
	if !ok {
		return
	}
	var s *model.Snapshot
	err := model.DB.Where("key = ? and bookmark_id = ?", sid, bid).First(&s).Error
	if err != nil {
		return
	}
	var b *model.Bookmark
	err = model.DB.Where("id = ?", bid).First(&b).Error
	if err != nil {
		setNotification(c, nError, err.Error(), false)
		return
	}
	if s.BookmarkID != b.ID {
		setNotification(c, nError, "Invalid bookmark ID", false)
		return
	}
	err = model.DB.Where("key = ? and bookmark_id = ?", sid, bid).First(&s).Error
	if err != nil {
		setNotification(c, nError, err.Error(), false)
		return
	}
	if s.Size == 0 {
		s.Size = storage.GetSnapshotSize(s.Key)
		err = model.DB.Save(s).Error
		if err != nil {
			setNotification(c, nError, err.Error(), false)
			return
		}
	}
	var otherSnapshots []struct {
		Title string
		Bid   int64
		Sid   string
	}
	err = model.DB.
		Model(&model.Snapshot{}).
		Select("bookmarks.id as bid, snapshots.key as sid, snapshots.title as title").
		Joins("join bookmarks on bookmarks.id = snapshots.bookmark_id").
		Where("bookmarks.url = ? and snapshots.key != ?", b.URL, s.Key).Find(&otherSnapshots).Error
	if err != nil {
		setNotification(c, nError, err.Error(), false)
		return
	}
	render(c, http.StatusOK, "snapshotWrapper", map[string]interface{}{
		"Bookmark":       b,
		"Snapshot":       s,
		"hideFooter":     true,
		"OtherSnapshots": otherSnapshots,
	})
}

func downloadSnapshot(c *gin.Context) {
	id, ok := c.GetQuery("sid")
	if !ok {
		return
	}
	r, err := storage.GetSnapshot(id)
	if err != nil {
		return
	}
	defer r.Close()
	gr, err := gzip.NewReader(r)
	if err != nil {
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=omnom_snapshot_%s.html;", id))
	c.Status(http.StatusOK)
	doc := html.NewTokenizer(gr)
	for {
		tt := doc.Next()
		switch tt {
		case html.ErrorToken:
			err := doc.Err()
			if errors.Is(err, io.EOF) {
				return
			}
			log.Println("Error: unexpected error while parsing")
			return
		case html.SelfClosingTagToken:
		case html.StartTagToken:
			if !writeBytes(c.Writer, []byte("<")) {
				return
			}
			tn, hasAttr := doc.TagName()
			if !writeBytes(c.Writer, tn) {
				return
			}
			if hasAttr {
				generateTagAttributes(tn, doc, c.Writer)
			}
			if !writeBytes(c.Writer, []byte(">")) {
				return
			}
		case html.TextToken:
			if !writeBytes(c.Writer, doc.Text()) {
				return
			}
		case html.EndTagToken:
			tn, _ := doc.TagName()
			if !writeBytes(c.Writer, []byte(fmt.Sprintf(`</%s>`, tn))) {
				return
			}
		}
	}
}

func generateTagAttributes(tagName []byte, doc *html.Tokenizer, out io.Writer) {
	for {
		aName, aVal, moreAttr := doc.TagAttr()
		if !writeBytes(out, []byte(fmt.Sprintf(` %s="`, aName))) {
			return
		}
		if attributeHasResource(tagName, aName, aVal) {
			href := string(aVal)
			res, err := storage.GetResource(filepath.Base(href))
			if err == nil {
				gres, err := gzip.NewReader(res)
				if err == nil {
					ext := filepath.Ext(href)
					if !writeBytes(out, []byte(fmt.Sprintf("data:%s;base64,", strings.Split(mime.TypeByExtension(ext), ";")[0]))) {
						return
					}
					bw := base64.NewEncoder(base64.StdEncoding, out)
					// TODO configure file size
					if _, err := io.CopyN(bw, gres, 1024*1024); err != nil && !errors.Is(err, io.EOF) {
						log.Println("Error: IO copy error", href, err)
						return
					}
					bw.Close()
					res.Close()
				}
			} else {
				if !writeBytes(out, aVal) {
					return
				}
			}
		} else {
			if !writeBytes(out, aVal) {
				return
			}
		}
		if !writeBytes(out, []byte(`"`)) {
			return
		}
		if !moreAttr {
			break
		}
	}
}

func attributeHasResource(tag, name, val []byte) bool {
	if resourceAttributes[string(tag)] == string(name) && bytes.HasPrefix(val, []byte("../../resources/")) {
		return true
	}
	return false
}

func writeBytes(w io.Writer, data []byte) bool {
	if l, err := w.Write(data); err != nil || l != len(data) {
		log.Println("Error: IO error")
		return false
	}
	return true
}

func deleteSnapshot(c *gin.Context) {
	u, _ := c.Get("user")
	session := sessions.Default(c)
	defer func() {
		_ = session.Save()
	}()
	bid := c.PostForm("bid")
	sid := c.PostForm("sid")
	if bid == "" || sid == "" {
		return
	}
	var s *model.Snapshot
	err := model.DB.
		Model(&model.Snapshot{}).
		Joins("join bookmarks on bookmarks.id = snapshots.bookmark_id").
		Where("snapshots.id = ? and snapshots.bookmark_id = ? and bookmarks.user_id", sid, bid, u.(*model.User).ID).First(&s).Error
	if err != nil {
		setNotification(c, nError, "Failed to delete snapshot: "+err.Error(), true)
	} else {
		setNotification(c, nInfo, "Snapshot deleted", true)
	}
	if s != nil {
		err = model.DB.Delete(&model.Snapshot{}, "id = ? and bookmark_id = ?", sid, bid).Error
		if err != nil {
			setNotification(c, nError, "Failed to delete snapshot: "+err.Error(), true)
		}
	}
	c.Redirect(http.StatusFound, baseURL("/edit_bookmark?id="+bid))
}
