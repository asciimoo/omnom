{{ define "content" }}
<div class="content">
  {{ if not .Bookmarks }}
  <h3 class="title">No bookmarks found</h3>
  {{ else }}
  <h3 class="title">Bookmarks</h3>
    {{ range .Bookmarks }}
    <div class="box media">
      <div class="media-content">
        <h4 class="title">{{ if .Favicon }}<img src="{{ .Favicon | ToURL }}" alt="favicon" /> {{ end }}<a href="{{ .URL }}" target="_blank">{{ .Title }}</a></h4>
        <p>{{ .Notes }}</p>
      </div>
      <div class="media-right">
          {{ range $i,$s := .Snapshots }}
            <a href="/snapshot?id={{ $s.ID }}">snapshot #{{ $i }}</a>
          {{ end }}
          {{ .CreatedAt | ToDate }} {{ if .Public }}Public{{ else }}Private{{ end }}
      </div>
    </div>
    {{ end}}
  {{ end }}
</div>
{{ end }}
