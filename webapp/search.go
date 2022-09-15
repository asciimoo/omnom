package webapp

import (
	"database/sql"
	"fmt"
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
	IsPublic         bool   `form:"public"`
	IsPrivate        bool   `form:"private"`
	SearchInSnapshot bool   `form:"search_in_snapshot"`
	SearchInNote     bool   `form:"search_in_note"`
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
