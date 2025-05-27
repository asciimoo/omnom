package webapp

import (
	"compress/gzip"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/asciimoo/omnom/contentdiff"
	"github.com/asciimoo/omnom/model"
	"github.com/asciimoo/omnom/storage"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func snapshotDiff(c *gin.Context) {
	s1, err := model.GetSnapshotWithResources(c.Query("s1"))
	if err != nil {
		render(c, http.StatusOK, "snapshot-diff-form", gin.H{})
		return
	}
	s2, err := model.GetSnapshotWithResources(c.Query("s2"))
	if err != nil {
		render(c, http.StatusOK, "snapshot-diff-form", gin.H{})
		return
	}

	var sURL string
	err = model.DB.
		Model(&model.Bookmark{}).
		Select("bookmarks.url").
		Joins("join snapshots on bookmarks.id = snapshots.bookmark_id").
		Where("snapshots.id == ?", s1.ID).First(&sURL).Error
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch URL for snapshot")
	}

	tds := contentdiff.DiffText(s1.Text, s2.Text)
	iKeys := contentdiff.DiffList(getImageResources(s1), getImageResources(s2))
	tdLen := 0
	for _, d := range tds {
		if d.Type != "0" {
			tdLen += 1
		}
	}

	sr1, err := createSnapshotReader(s1.Key)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create snapshot reader")
		setNotification(c, nError, "Backend error", false)
		render(c, http.StatusOK, "snapshot-diff-form", gin.H{})
		return
	}
	sr2, err := createSnapshotReader(s2.Key)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create snapshot reader")
		setNotification(c, nError, "Backend error", false)
		render(c, http.StatusOK, "snapshot-diff-form", gin.H{})
		return
	}
	c1 := contentdiff.ExtractHTMLContent(sr1)
	c2 := contentdiff.ExtractHTMLContent(sr2)
	render(c, http.StatusOK, "snapshot-diff", gin.H{
		"TextDiffs":   tds,
		"TextDiffLen": tdLen,
		"ImageDiffs":  iKeys,
		"LinkDiffs":   contentdiff.DiffLink(c1.Links, c2.Links),
		"SURL":        sURL,
		"S1":          s1,
		"S2":          s2,
	})
}

func snapshotDiffSideBySide(c *gin.Context) {
	s1, err := model.GetSnapshotWithResources(c.Query("s1"))
	if err != nil {
		notFoundView(c)
		return
	}
	s2, err := model.GetSnapshotWithResources(c.Query("s2"))
	if err != nil {
		notFoundView(c)
		return
	}
	var sURL string
	err = model.DB.
		Model(&model.Bookmark{}).
		Select("bookmarks.url").
		Joins("join snapshots on bookmarks.id = snapshots.bookmark_id").
		Where("snapshots.id == ?", s1.ID).First(&sURL).Error
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch URL for snapshot")
	}

	tds := contentdiff.DiffText(s1.Text, s2.Text)
	type Edit struct {
		S    string `json:"s"`
		Idx  int    `json:"idx"`
		PreS string `json:"preStr"`
	}
	adds := make([]Edit, 0, 16)
	dels := make([]Edit, 0, 16)
	var b strings.Builder
	for _, d := range tds {
		switch d.Type {
		case "+":
			idx := strings.Count(b.String(), d.Text) + 1
			adds = append(adds, Edit{S: d.Text, Idx: idx})
		case "-":
			idx := 1
			var preS string
			if len(b.String()) < 7 {
				preS = b.String()
			} else {
				preS := b.String()[b.Len()-5:]
				idx = strings.Count(b.String(), preS) + 1
			}
			dels = append(dels, Edit{S: d.Text, Idx: idx, PreS: preS})
		}
		if d.Type == "+" || d.Type == "0" {
			b.WriteString(d.Text)
		}
	}
	as, err := json.Marshal(adds)
	if err != nil {
		log.Error().Err(err).Msg("Failed to serialize additions")
		as = []byte("[]")
	}
	ds, err := json.Marshal(dels)
	if err != nil {
		log.Error().Err(err).Msg("Failed to serialize deletions")
		ds = []byte("[]")
	}
	render(c, http.StatusOK, "snapshot-diff-side-by-side", gin.H{
		"SURL":       sURL,
		"S1":         s1,
		"S2":         s2,
		"hideFooter": true,
		//nolint: gosec // JSON is safe.
		"Additions": template.JS(string(as)),
		//nolint: gosec // JSON is safe.
		"Deletions": template.JS(string(ds)),
	})
}

func getImageResources(s *model.Snapshot) []string {
	ret := make([]string, 0, 8)
	for _, r := range s.Resources {
		if strings.HasPrefix(strings.ToLower(r.MimeType), "image") {
			ret = append(ret, r.Key)
		}
	}
	return ret
}

func createSnapshotReader(sid string) (io.ReadCloser, error) {
	r, err := storage.GetSnapshot(sid)
	if err != nil {
		return nil, err
	}
	return gzip.NewReader(r)
}
