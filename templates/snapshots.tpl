{{ define "content" }}
<div class="content">
    <div class="is-pulled-right"><a href="{{ AddURLParam .URL "format=rss" }}">RSS<span class="icon"><i class="fas fa-rss"></i></span></a></div>
    <h3 class="title">Find snapshots by domain or URL</h3>
    {{ block "info" "Check out the search functionality on the Bookmark listing for full text search and advanced filtering" }}{{ end }}
    <form action="" method="get">
    <div class="columns">
        <div class="column">
        {{ block "textFilter" .}}{{ end }}
        {{ block "submit" "Search" }}{{ end }}
        </div>
    </div>
    </form>
    {{ if .SearchParams.Q }}
        {{ if eq .SnapshotCount 0 }}
            <h3 class="title">No snapshots found</h3>
        {{ else }}
            <h3 class="title">Results for "{{ .SearchParams.Q }}" ({{ .SnapshotCount }})</h3>
            {{ range .Snapshots }}
            <div class="box">
                <h4 class="title"><a href="{{ URLFor "Snapshot" }}?sid={{ .Key }}&bid={{ .BookmarkID }}">{{ .Bookmark.Title }}</a></h4>
                <p>
                    Original URL: <a href="{{ .Bookmark.URL }}" target="_blank">{{ Truncate .Bookmark.URL 100 }}</a><br />
                    {{ .UpdatedAt | ToDateTime }} - {{ .Size | FormatSize }}
                </p>
            </div>
            {{ end }}
        {{ end }}
    {{ block "paging" .}}{{ end }}
    {{ end }}
</div>
{{ end }}
