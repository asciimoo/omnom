{{ define "content" }}
<div class="content">
    <h2 class="title">{{ .User.Username }} <span class="is-size-7 is-italic has-text-grey">{{ .User.Email }}</span></h2>
    {{ if not .AddonTokens }}
    <h3 class="title">No addon token found</h3>
    {{ else }}
    <h3 class="title">Addon tokens</h3>
    <div class="columns is-mobile"><div class="column is-narrow">
        <div class="list"><dl>
            {{ range .AddonTokens }}
            <div class="list-item"><li class="pure-list"><code class="has-text-dark">{{ .Text }}</code> <a href="/delete_addon_token?id={{ .ID }}" class="ml-3" title="delete token"><span class="icon has-text-danger"><i class="fas fa-trash-alt"></i></span></a></li></div>
            {{ end}}
        </dl></div>
    </div></div>
    {{ end }}
    <a href="/generate_addon_token" class="button is-primary">Generate new addon token</a>
</div>
{{ end }}
