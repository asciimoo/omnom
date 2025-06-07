{{ define "content" }}
<div class="content">
    <h2 class="title is-size-2 mt-3 mb-0">{{ .Bookmark.Title }}</h2>
    <p class="icon-text mb-0">
        {{ if .Bookmark.Favicon }}
        <span class="icon">
            <img src="{{ .Bookmark.Favicon | ToURL }}" alt="favicon" />
        </span>
        {{ end }}
        <span class="is-size-6 has-text-grey has-text-weight-normal">{{ Truncate .Bookmark.URL 200 }}</span>
    </p>
    <p class="mt-2">
        {{ $uid := 0 }}
        {{ if .User }}{{ $uid = .User.ID }}{{ end }}
        {{ if .Bookmark.Tags }}
            {{ range .Bookmark.Tags }}
            <a href="{{ if ne $uid $.Bookmark.UserID }}{{ BaseURL "/bookmarks" }}{{ else }}{{ BaseURL "/my_bookmarks" }}{{ end }}?tag={{ .Text }}"><span class="tag is-muted-primary is-medium">{{ .Text }}</span></a>
            {{ end }}
            <br />
        {{ end }}
        <span>{{ .Bookmark.CreatedAt | ToDateTime }} - {{ if .Bookmark.Public }}Public{{ else }}Private{{ end }}</span>
        {{ if .User }}
        {{ if eq .User.ID .Bookmark.UserID }}
            <br /><span> <a href="{{ BaseURL "/edit_bookmark" }}?id={{ .Bookmark.ID }}">edit</a></span>
        {{ end }}
        {{ end }}
    </p>
    {{ if .Bookmark.Notes }}
        <h4>Notes</h4>
        <p>{{ .Bookmark.Notes }}</p>
    {{ end }}
    {{ if .Bookmark.Snapshots }}
        <div class="mt-6">
            <h4>Snapshots</h4>
            {{ block "snapshots" KVData "Snapshots" .Bookmark.Snapshots "IsOwn" (eq .Bookmark.UserID $uid ) }}{{ end }}
        </div>
    {{ end }}
</div>
{{ end }}
