{{ define "content" }}
<h2 class="title">API documentation</h2>
<h3 class="title is-size-3">Endpoints</h3>
{{ range .Endpoints }}
<div class="box media">
    <div class="media-content">
        <h4 class="title"><code class="has-background-danger-light">{{ .Method }}</code><code>{{ .Path }}</code></h4>
        <h4 class="title">{{ .Name }}</h4>
        <p>{{ .Description }}</p>
    </div>
</div>
{{ end }}
{{ end }}
