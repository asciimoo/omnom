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
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/html"

	"github.com/asciimoo/omnom/model"
	"github.com/asciimoo/omnom/storage"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
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
		Title     string
		Bid       int64
		Key       string
		CreatedAt time.Time
	}
	err = model.DB.
		Model(&model.Snapshot{}).
		Select("bookmarks.id as bid, snapshots.key as key, snapshots.title as title, snapshots.created_at as created_at").
		Joins("join bookmarks on bookmarks.id = snapshots.bookmark_id").
		Where("bookmarks.url = ? and snapshots.key != ?", b.URL, s.Key).Find(&otherSnapshots).Error
	if err != nil {
		setNotification(c, nError, err.Error(), false)
		return
	}
	render(c, http.StatusOK, "snapshot-wrapper", map[string]interface{}{
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
	if !writeBytes(c.Writer, []byte("<!DOCTYPE html>")) {
		return
	}
	doc := html.NewTokenizer(gr)
	for {
		tt := doc.Next()
		switch tt {
		case html.ErrorToken:
			err := doc.Err()
			if errors.Is(err, io.EOF) {
				return
			}
			log.Error().Err(err).Msg("Failed to parse HTML")
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
				defer res.Close()
				gres, err := gzip.NewReader(res)
				if err == nil {
					ext := filepath.Ext(href)
					if !writeBytes(out, []byte(fmt.Sprintf("data:%s;base64,", strings.Split(mime.TypeByExtension(ext), ";")[0]))) {
						return
					}
					if ext == ".css" {
						err = writeB64(generateCSS(gres), out)
					} else {
						err = writeB64(gres, out)
					}
					if err != nil {
						log.Error().Err(err).Str("href", href).Msg("IO error")
					}
				}
			} else {
				if !writeBytes(out, aVal) {
					return
				}
			}
		} else {
			if !writeBytes(out, []byte(html.EscapeString(string(aVal)))) {
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

func generateCSS(in io.Reader) io.Reader {
	out := bytes.NewBuffer(nil)
	l := css.NewLexer(parse.NewInput(in))
	for {
		tt, text := l.Next()
		switch tt {
		case css.ErrorToken:
			return out
		case css.URLToken:
			href := string(text[5 : len(text)-2])
			if strings.HasPrefix(href, "../../resources/") {
				ext := filepath.Ext(href)
				res, err := storage.GetResource(filepath.Base(href))
				if err == nil {
					defer res.Close()
					writeBytes(out, []byte(fmt.Sprintf("url(\"data:%s;base64,", strings.Split(mime.TypeByExtension(ext), ";")[0])))
					err := writeB64(res, out)
					if err != nil {
						log.Error().Err(err).Msg("CSS IO error")
					}
					writeBytes(out, []byte("\")"))
				} else {
					out.Write(text)
				}
			} else {
				out.Write(text)
			}
		// TODO handle CSS @import
		default:
			out.Write(text)
		}
	}
	return out
}

func writeB64(in io.Reader, out io.Writer) error {
	bw := base64.NewEncoder(base64.StdEncoding, out)
	// TODO configure file size
	if _, err := io.CopyN(bw, in, 1024*1024); err != nil && !errors.Is(err, io.EOF) {
		log.Error().Err(err).Msg("IO copy error")
		return err
	}
	defer bw.Close()
	return nil
}

func attributeHasResource(tag, name, val []byte) bool {
	if resourceAttributes[string(tag)] == string(name) && bytes.HasPrefix(val, []byte("../../resources/")) {
		return true
	}
	return false
}

func writeBytes(w io.Writer, data []byte) bool {
	if l, err := w.Write(data); err != nil || l != len(data) {
		log.Error().Err(err).Msg("IO error")
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

func snapshots(c *gin.Context) {
	qs, ok := c.GetQuery("query")
	if !ok {
		render(c, http.StatusOK, "snapshots", nil)
		return
	}
	var uid uint = 0
	u, ok := c.Get("user")
	if ok && u != nil {
		uid = u.(*model.User).ID
	}
	var ss []*model.Snapshot
	pageno := getPageno(c)
	offset := (pageno - 1) * resultsPerPage
	var sc int64
	//cq := model.DB.Model(&model.Snapshot{}).Where("s.public = 1")
	//nolint: gosec // uint -> int conversion is safe
	q := model.DB.Limit(int(resultsPerPage)).Offset(int(offset)).Joins("left join bookmarks on bookmarks.id = snapshots.bookmark_id").Where("bookmarks.url like ?", "%"+qs+"%").Where("bookmarks.public == true or bookmarks.user_id == ?", uid).Preload("Bookmark")
	cq := model.DB.Model(&model.Snapshot{}).Joins("left join bookmarks on bookmarks.id = snapshots.bookmark_id").Where("bookmarks.url like ?", "%"+qs+"%").Where("bookmarks.public == true or bookmarks.user_id == ?", uid)
	cq.Count(&sc)
	q.Order("snapshots.created_at").Find(&ss)
	render(c, http.StatusOK, "snapshots", map[string]interface{}{
		"Snapshots":     ss,
		"SnapshotCount": sc,
		"Pageno":        pageno,
		//nolint: gosec // uint -> int conversion is safe
		"HasNextPage": offset+resultsPerPage < uint(sc),
		"SearchParams": searchParams{
			Q: qs,
		},
	})
}

func snapshotDetails(c *gin.Context) {
	sid, ok := c.GetQuery("sid")
	if !ok {
		return
	}
	var s *model.Snapshot
	err := model.DB.Where("key = ?", sid).First(&s).Error
	if err != nil {
		return
	}
	var b *model.Bookmark
	err = model.DB.Where("id = ?", s.BookmarkID).First(&b).Error
	if err != nil {
		return
	}
	var res []*model.Resource
	err = model.DB.Model(&model.Resource{}).
		Joins("join snapshot_resources as sr on sr.resource_id == resources.id").
		Joins("join snapshots on sr.snapshot_id == snapshots.id").
		Where("snapshots.key = ?", sid).Find(&res).Error
	if err != nil {
		return
	}
	rs := make(map[string]map[string][]*model.Resource)
	for _, v := range res {
		if strings.TrimSpace(v.OriginalFilename) == "" {
			continue
		}
		m, _, err := mime.ParseMediaType(v.MimeType)
		if err != nil {
			continue
		}
		mParts := strings.Split(m, "/")
		mType := mParts[0]
		mSubtype := "unknown"
		if len(mParts) > 1 {
			mSubtype = mParts[1]
		}
		if _, ok := rs[mType]; !ok {
			rs[mType] = make(map[string][]*model.Resource)
		}
		if _, ok := rs[mType][mSubtype]; !ok {
			rs[mType][mSubtype] = make([]*model.Resource, 0, 4)
		}
		rs[mType][mSubtype] = append(rs[mType][mSubtype], v)
	}
	render(c, http.StatusOK, "snapshot-details", map[string]interface{}{
		"Snapshot":      s,
		"Resources":     rs,
		"ResourceCount": len(res),
		"URL":           b.URL,
		"Bookmark":      b,
	})
}
