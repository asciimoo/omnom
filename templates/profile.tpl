{{ define "content" }}
<div class="content">
    <h2 class="title">{{ .User.Username }} <span class="is-size-7 is-italic has-text-grey">{{ .User.Email }}</span></h2>
    <p>Snapshot storage usage: <strong>{{ .SnapshotsSize | FormatSize }}</strong></p>
    {{ if not .AddonTokens }}
    <h3 class="title">No addon token found</h3>
    {{ else }}
    <h3 class="title">Addon tokens</h3>
    <div class="columns is-mobile"><div class="column is-narrow">
        <div class="list"><dl>
            {{ range .AddonTokens }}
            <div class="list-item">
                <li class="pure-list">
                    <form method="post" action="{{ BaseURL "/delete_addon_token" }}">
                            <code class="has-text-dark">{{ .Text }}</code>
                            <input type="hidden" name="_csrf" value="{{ $.CSRF }}" />
                            <input type="hidden" name="id" value="{{ .ID }}" />
                            <input type="submit" class="button is-danger is-small" value="Delete" />
                    </form>
                </li>
            </div>
            {{ end}}
        </dl></div>
    </div></div>
    {{ end }}
    <a href="{{ BaseURL "/generate_addon_token" }}" class="button is-primary">Generate new addon token</a>
</div>
{{ end }}
