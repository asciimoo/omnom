{{ define "full-content" }}
<div class="snapshot__container">
    <div class="container content my-5">
        <h3 class="title mb-0">Snapshot of <a href="{{ .Bookmark.URL }}">{{ Truncate .Bookmark.URL 100 }}</a></h3>
        <p>
            <strong>{{ .Snapshot.CreatedAt | ToDate }}</strong>
            <span class="tag is-info is-light">{{ .Snapshot.Size | FormatSize }}</span> <a href="{{ SnapshotURL .Snapshot.Key }}"><small>Fullscreen</small></a>
            - <a href="{{ BaseURL "/download_snapshot" }}?sid={{ .Snapshot.Key }}"><small>Download</small></a>
        </p>
    </div>
    {{ if .OtherSnapshots }}
    <div class="accordion-tabs">
        <div class="accordion-tab">
            <input class="accordion-tab__control" type="checkbox" id="chck2">
            <label class="accordion-tab-label" for="chck2">
                <div class="my-bookmarks__section-header">
                    <h3>
                        Other snapshots of this URL
                    </h3>
                    <i class="fas fa-angle-down"></i>
                </div>
            </label>
            <div class="accordion-tab-content">
                {{ range $i,$s := .OtherSnapshots }}
                <span class="tag"><a href="{{ BaseURL "/snapshot" }}?sid={{ $s.Sid }}&bid={{ $s.Bid }}">{{ if $s.Title }}{{ $s.Title }}{{ else }}#{{ $i }}{{ end }}</a></span>
                {{ end }}
            </div>
        </div>
    </div>
    {{ end }}
</div>
<iframe src="{{ SnapshotURL .Snapshot.Key }}" title="snapshot of {{ .Bookmark.URL }}"></iframe>
{{ end }}
