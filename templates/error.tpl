{{ define "content" }}
<div class="content">
    <h2 class="title">Error! {{ .Title }}</h2>
    {{ if .Message }}<p>{{ .Message }}</p>{{ end }}
</div>
{{ end }}
