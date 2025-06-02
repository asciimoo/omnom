// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package webapp

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/asciimoo/omnom/model"

	"gorm.io/gorm"
)

const dateFormat string = "2006.01.02"

type searchParams struct {
	Q                string `form:"query"`
	Owner            string `form:"owner"`
	FromDate         string `form:"from"`
	ToDate           string `form:"to"`
	Tag              string `form:"tag"`
	Domain           string `form:"domain"`
	Collection       string `form:"collection"`
	IsPublic         bool   `form:"public"`
	IsPrivate        bool   `form:"private"`
	SearchInSnapshot bool   `form:"search_in_snapshot"`
	SearchInNote     bool   `form:"search_in_note"`
}

func (s *searchParams) Serialize() string {
	v := url.Values{}
	v.Add("query", s.Q)
	v.Add("user", s.Owner)
	v.Add("from", s.FromDate)
	v.Add("to", s.ToDate)
	v.Add("tag", s.Tag)
	v.Add("domain", s.Domain)
	v.Add("collection", s.Collection)
	if s.IsPublic {
		v.Add("public", "1")
	}
	if s.IsPrivate {
		v.Add("private", "1")
	}
	if s.SearchInSnapshot {
		v.Add("search_in_snapshot", "1")
	}
	if s.SearchInNote {
		v.Add("search_in_note", "1")
	}
	return v.Encode()
}

func (s *searchParams) String() string {
	parts := make([]string, 0, 4)
	if s.Q != "" {
		parts = append(parts, ".q_"+s.Q)
	}
	if s.Owner != "" {
		parts = append(parts, ".u_"+s.Owner)
	}
	if s.FromDate != "" {
		parts = append(parts, ".from_"+s.FromDate)
	}
	if s.ToDate != "" {
		parts = append(parts, ".to_"+s.ToDate)
	}
	if s.Tag != "" {
		parts = append(parts, ".t_"+s.Tag)
	}
	if s.Domain != "" {
		parts = append(parts, ".d_"+s.Domain)
	}
	if s.IsPublic {
		parts = append(parts, ".public")
	}
	if s.IsPrivate {
		parts = append(parts, ".private")
	}
	if s.SearchInSnapshot {
		parts = append(parts, ".snapshot")
	}
	if s.SearchInNote {
		parts = append(parts, ".note")
	}
	return strings.Join(parts, "")
}

func filterText(qs string, inNote bool, inSnapshot bool, q, cq *gorm.DB) {
	if qs == "" {
		return
	}
	if strings.Contains(qs, "*") {
		qs = strings.ReplaceAll(qs, "*", "%")
	} else {
		qs = fmt.Sprintf("%%%s%%", qs)
	}
	query := "LOWER(bookmarks.title) LIKE LOWER(@query)"
	if inNote {
		query += " or LOWER(bookmarks.notes) LIKE LOWER(@query)"
	}
	if inSnapshot {
		q = q.Joins("join snapshots on snapshots.bookmark_id = bookmarks.id")
		cq = cq.Joins("join snapshots on snapshots.bookmark_id = bookmarks.id")
		query += " or LOWER(snapshots.text) LIKE LOWER(@query)"
	}
	query = "(" + query + ")"
	q = q.Where(query, sql.Named("query", qs))   //nolint: staticcheck,wastedassign // it is used in later funcs
	cq = cq.Where(query, sql.Named("query", qs)) //nolint: staticcheck,wastedassign // it is used in later funcs
}

func filterCollection(cid string, uid uint, q, cq *gorm.DB) {
	if cid == "" {
		return
	}
	q = q. //nolint: staticcheck,wastedassign // it is used in later funcs
		Joins("join collections on bookmarks.collection_id == collections.id").
		Where("collections.name = ?", cid).
		Where("collections.user_id = ? ", uid)
	cq = cq. //nolint: staticcheck,wastedassign // it is used in later funcs
			Joins("join collections on bookmarks.collection_id == collections.id").
			Where("collections.name = ?", cid).
			Where("collections.user_id = ? ", uid)
}

func filterOwner(o string, q, cq *gorm.DB) {
	if o == "" {
		return
	}
	u := model.GetUser(o)
	if u == nil {
		return
	}
	q = q.Where("user_id == ? and public == true", u.ID)   //nolint: staticcheck,wastedassign // it is used in later funcs
	cq = cq.Where("user_id == ? and public == true", u.ID) //nolint: staticcheck,wastedassign // it is used in later funcs
}

func filterDomain(d string, q, cq *gorm.DB) {
	if d == "" {
		return
	}
	q = q.Where("domain LIKE ?", fmt.Sprintf("%%%s%%", d))   //nolint: staticcheck,wastedassign // it is used in later funcs
	cq = cq.Where("domain LIKE ?", fmt.Sprintf("%%%s%%", d)) //nolint: staticcheck,wastedassign // it is used in later funcs
}

func filterTag(t string, q, cq *gorm.DB) {
	if t == "" {
		return
	}
	q = q. //nolint: staticcheck,wastedassign // it is used in later funcs
		Joins("join bookmark_tags on bookmark_tags.bookmark_id == bookmarks.id").
		Joins("join tags on bookmark_tags.tag_id == tags.id").
		Where("tags.text = ?", t)
	cq = cq. //nolint: staticcheck,wastedassign // it is used in later funcs
			Joins("join bookmark_tags on bookmark_tags.bookmark_id == bookmarks.id").
			Joins("join tags on bookmark_tags.tag_id == tags.id").
			Where("tags.text = ?", t)
}

func filterFromDate(d string, q, cq *gorm.DB) error {
	if d == "" {
		return nil
	}
	t, err := time.Parse(dateFormat, d)
	if err != nil {
		return err
	}
	q = q.Where("bookmarks.created_at >= ?", t)   //nolint: staticcheck,wastedassign // it is used in later funcs
	cq = cq.Where("bookmarks.created_at >= ?", t) //nolint: staticcheck,wastedassign // it is used in later funcs
	return nil
}

func filterToDate(d string, q, cq *gorm.DB) error {
	if d == "" {
		return nil
	}
	t, err := time.Parse(dateFormat, d)
	if err != nil {
		return err
	}
	q = q.Where("bookmarks.created_at <= ?", t)   //nolint: staticcheck,wastedassign // it is used in later funcs
	cq = cq.Where("bookmarks.created_at <= ?", t) //nolint: staticcheck,wastedassign // it is used in later funcs
	return nil
}

func filterPublic(q, cq *gorm.DB) {
	q.Where("public == true")
	cq.Where("public == true")
}
