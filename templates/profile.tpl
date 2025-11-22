{{ define "content" }}
<div class="content">
    <h2 class="title">{{ .User.Username }}
	<span class="is-size-7 is-italic has-text-grey">{{ or .User.Email (.Tr.Msg "no email provided") }}</span></h2>
    <p><a href="{{ URLFor "Logout" }}" class="button is-warning">{{ .Tr.Msg "logout" }}</a></p>
    <p>{{ .Tr.Msg "storage usage" }}: <strong>{{ .SnapshotsSize | FormatSize }}</strong></p>
    {{ if .DisplayTokens }}
        {{ if not .AddonTokens }}
        <h3 class="title">{{ .Tr.Msg "no addon token found" }}</h3>
        {{ else }}
        <h3 class="title">{{ .Tr.Msg "addon tokens" }}</h3>
        <div class="columns is-mobile"><div class="column is-narrow">
            <div class="list"><dl>
                {{ $Tr := .Tr }}
                {{ range .AddonTokens }}
                <div class="list-item">
                    <li class="pure-list">
                        <form method="post" action="{{ URLFor "Delete addon token" }}">
                                <code class="has-text-dark">{{ .Text }}</code>
                                <input type="hidden" name="id" value="{{ .ID }}" />
                                <input type="submit" class="button is-danger is-small" value="{{ $Tr.Msg "delete" }}" />
                        </form>
                    </li>
                </div>
                {{ end}}
            </dl></div>
        </div></div>
        {{ end }}
    {{ else }}
        <div class="columns is-mobile"><div class="column is-narrow">
            <form method="post">
                <input type="submit" class="button is-primary" value="{{ .Tr.Msg "show addon tokens" }}" />
            </form>
        </div></div>
    {{ end }}
    <a href="{{ URLFor "Generate addon token" }}" class="button is-primary">{{ .Tr.Msg "generate addon token" }}</a>
</div>
{{ end }}
