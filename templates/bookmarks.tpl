{{ define "content" }}
<div class="content">
{{ if not .Bookmarks }}
  <h3 class="title">No public bookmarks found</h3>
{{ else }}
  <h3 class="title">Public bookmarks ({{ .BookmarkCount }})</h3>
  {{ range .Bookmarks }}
    {{ block "bookmark" .}}{{ end }}
  {{ end}}
{{ end }}
{{ block "paging" .}}{{ end }}
</div>
{{ end }}
