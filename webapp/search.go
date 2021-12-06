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

func filterText(qs string, inNote bool, inSnapshot bool, q, cq *gorm.DB) error {
	if qs == "" {
		return nil
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
	q = q.Where(query, sql.Named("query", qs))
	cq = cq.Where(query, sql.Named("query", qs))
	return nil
}

func filterOwner(o string, q, cq *gorm.DB) error {
	if o == "" {
		return nil
	}
	u := model.GetUser(o)
	if u == nil {
		return nil
	}
	q = q.Where("user_id == ? and public == true", u.ID)
	cq = cq.Where("user_id == ? and public == true", u.ID)
	return nil
}

func filterDomain(d string, q, cq *gorm.DB) error {
	if d == "" {
		return nil
	}
	q = q.Where("domain LIKE ?", fmt.Sprintf("%%%s%%", d))
	cq = cq.Where("domain LIKE ?", fmt.Sprintf("%%%s%%", d))
	return nil
}

func filterTag(t string, q, cq *gorm.DB) error {
	if t == "" {
		return nil
	}
	q = q.Joins("join tags on tags.bookmark_id == bookmarks.id").Where("tags.text = ?", t)
	cq = cq.Joins("join tags on tags.bookmark_id == bookmarks.id").Where("tags.text = ?", t)
	return nil
}

func filterFromDate(d string, q, cq *gorm.DB) error {
	if d == "" {
		return nil
	}
	t, err := time.Parse(dateFormat, d)
	if err != nil {
		return err
	}
	q = q.Where("bookmarks.created_at >= ?", t)
	cq = cq.Where("bookmarks.created_at >= ?", t)
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
	q = q.Where("bookmarks.created_at <= ?", t)
	cq = cq.Where("bookmarks.created_at <= ?", t)
	return nil
}

func filterPublic(q, cq *gorm.DB) error {
	q.Where("public == true")
	cq.Where("public == true")
	return nil
}

func filterPrivate(q, cq *gorm.DB) error {
	q.Where("public == false")
	cq.Where("public == false")
	return nil
}
