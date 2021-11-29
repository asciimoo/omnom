{{ define "content" }}

    {{ if not .Bookmarks }}
    <h3 class="title">No bookmarks found</h3>
    {{ else }}
    <h3 class="title">Search results{{ if .SearchParams.Q }} for "{{ .SearchParams.Q }}"{{ end }} ({{ .BookmarkCount }})</h3>
      {{ range .Bookmarks }}
        {{ block "bookmark" .}}{{ end }}
      {{ end}}
    {{ end }}
    {{ block "paging" .}}{{ end }}
</form></div>
{{ end }}
