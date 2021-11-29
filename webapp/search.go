package webapp

import (
	"fmt"

	"github.com/asciimoo/omnom/model"

	"gorm.io/gorm"
)

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
	q = q.Where("LOWER(title) LIKE LOWER(?)", fmt.Sprintf("%%%s%%", qs))
	cq = cq.Where("LOWER(title) LIKE LOWER(?)", fmt.Sprintf("%%%s%%", qs))
	if inNote {
		q = q.Or("LOWER(notes) LIKE LOWER(?)", fmt.Sprintf("%%%s%%", qs))
		cq = cq.Or("LOWER(notes) LIKE LOWER(?)", fmt.Sprintf("%%%s%%", qs))
	}
	if inSnapshot {
		// TODO
		fmt.Println(inSnapshot)
	}
	return nil
}

func filterOwner(o string, q, cq *gorm.DB) error {
	u := model.GetUser(o)
	if u == nil {
		return nil
	}
	q = q.Where("user_id == ? or public == true", u.ID)
	cq = cq.Where("user_id == ? or public == true", u.ID)
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

func filterFromDate(d string, q, cq *gorm.DB) error {
	// TODO
	fmt.Println(d, q, cq)
	return nil
}

func filterToDate(d string, q, cq *gorm.DB) error {
	// TODO
	fmt.Println(d, q, cq)
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
