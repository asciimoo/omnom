{{ define "content" }}
<div class="snapshot__container">
    <div class="content">
        <h3 class="title mb-0">Snapshot of <a href="{{ .Bookmark.URL }}">{{ Truncate .Bookmark.URL 100 }}</a></h3>
        <p><strong>{{ .Snapshot.CreatedAt | ToDate }}</strong> <span class="tag is-info is-light">{{ .Snapshot.Size | FormatSize }}</span> <a href="{{ BaseURL "/view_snapshot" }}?id={{ .Snapshot.Key }}"><small>Fullscreen snapshot</small></a></p>
    </div>
    <div class="iframe-box">
        <div class="iframe-container">
            <iframe src="{{ BaseURL "/view_snapshot" }}?id={{ .Snapshot.Key }}" title="snapshot of {{ .Bookmark.URL }}" width="100%" height="100%" frameborder="1px"></iframe>
        </div>
    </div>
</div>
{{ end }}
