{{ define "content" }}
<div class="content">
    <div class="is-pulled-right"><a href="{{ AddURLParam .URL "format=rss" }}">RSS<span class="icon"><i class="fas fa-rss"></i></span></a></div>
    <h3 class="title">{{ .Tr.Msg "find snapshots by domain or url" }}</h3>
    {{ block "info" KVData "Info" (.Tr.Msg "filtering tip") }}{{ end }}
    <form action="" method="get">
    <div class="columns">
        <div class="column">
        {{ block "textFilter" .}}{{ end }}
        {{ block "submit" (.Tr.Msg "search") }}{{ end }}
        </div>
    </div>
    </form>
    {{ if .SearchParams.Q }}
        {{ if eq .SnapshotCount 0 }}
            <h3 class="title">{{ .Tr.Msg "no snapshot found" }}</h3>
        {{ else }}
            <h3 class="title">Results for "{{ .SearchParams.Q }}" ({{ .SnapshotCount }})</h3>
            {{ $Tr := .Tr }}
            {{ range .Snapshots }}
            <div class="box">
                <h4 class="title"><a href="{{ URLFor "Snapshot" }}?sid={{ .Key }}&bid={{ .BookmarkID }}">{{ .Bookmark.Title }}</a></h4>
                <p>
                    {{ $Tr.Msg "original url" }}: <a href="{{ .Bookmark.URL }}" target="_blank">{{ Truncate .Bookmark.URL 100 }}</a><br />
                    {{ .UpdatedAt | ToDateTime }} - {{ .Size | FormatSize }}
                </p>
            </div>
            {{ end }}
        {{ end }}
    {{ block "paging" .}}{{ end }}
    {{ end }}
</div>
{{ end }}
