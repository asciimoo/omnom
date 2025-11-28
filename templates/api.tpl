{{ define "full-content" }}
<div class="fullscreen-wrapper">
    <h2 class="title">API documentation</h2>
    <div class="columns">
        <div class="column is-2-fullhd is-one-quarter-desktop is-one-third-tablet">
            <h3 class="title is-size-4">Endpoints</h3>
            <div class="content">
                <ul>
                    {{ range .Endpoints }}
                    <li><a href="#{{ Replace .Name " " "_" | ToLower}}_{{ .Method }}">{{ .Name }}</a></li>
                    {{ end }}
                </ul>
            </div>
        </div>
        <div class="column container">
            {{ range .Endpoints }}
            <div class="box media" id="{{ Replace .Name " " "_" | ToLower}}_{{ .Method }}">
                <div class="media-content">
                    <h3 class="title"><code class="has-background-danger-light">{{ .Method }}</code><code>{{ .Path }}</code></h3>
                    <h4 class="title is-size-5">{{ .Name }}{{ if .AuthRequired }}<span class="tag is-warning is-light is-size-7 has-text-weight-normal">authentication required</span>{{ end }}</h4>
                    <p>{{ .Description }}{{ if .RSS }}<br>Add <code>?format=rss</code> for RSS output{{ end }}</p>
                    <hr />
                    {{ if .Args }}
                    <h5 class="title is-size-6">Arguments</h5>
                    <table class="table is-bordered">
                        <tr>
                            <th>Name</th>
                            <th>Type</th>
                            <th>Required</th>
                            <th>Description</th>
                        </tr>
                        {{ range .Args }}
                        <tr>
                            <td><code>{{ .Name }}</code></td>
                            <td><code>{{ .Type }}</code></td>
                            <td>{{ .Required }}</td>
                            <td>{{ .Description }}</td>
                        {{ end }}
                    </table>
                    {{ else }}
                    <h5 class="title is-size-6">No arguments available for this endpoint</h5>
                    {{ end }}
                </div>
            </div>
            {{ end }}
        </div>
    </div>
</div>
{{ end }}
