{{ define "full-content" }}
<div class="fullscreen-wrapper">
    <h3 class="title">{{ .Tr.Msg "feeds" }}</h3>
    <div class="columns">
        <div class="column is-2-fullhd is-one-quarter-desktop is-one-third-tablet">
            <div class="content">
                {{ if not .Feeds }}
                <h3 class="title">{{ .Tr.Msg "no feeds found" }}</h3>
                {{ end }}
                <form action="" method="get">
                    {{ block "textFilter" .}}{{ end }}
                    {{ block "submit" (.Tr.Msg "search") }}{{ end }}
                </form>
                <details class="my-4 is-size-4">
                    <summary>{{ .Tr.Msg "add feed" }}</summary>
                    <form action="{{ URLFor "add feed" }}" method="post">
                        <div class="field">
                            <label class="label">{{ .Tr.Msg "name" }}</label>
                            <div class="control">
                                <input class="input" type="text" placeholder="{{ .Tr.Msg "name" }}.." name="name" />
                            </div>
                        </div>
                        <div class="field">
                            <label class="label">{{ .Tr.Msg "url" }}</label>
                            <div class="control">
                                <input class="input" type="text" placeholder="{{ .Tr.Msg "url" }}.." name="url" />
                            </div>
                        </div>
                        {{ block "submit" (.Tr.Msg "submit") }}{{ end }}
                    </form>
                </details>
                {{ range .Feeds }}
                <h4>{{ .Name }}</h4>
                {{ end }}
            </div>
        </div>
        {{ $uid := 0 }}
        {{ $page := .Page }}
        {{ $Tr := .Tr }}
        <div class="column">
            {{ if .UnreadItemCount }}
                <h3 class="title">{{ .Tr.Msg "unread items" }} ({{ .UnreadItemCount }})</h3>
                {{ range .UnreadItems }}
                    <div class="media">
                    <div class="media-left">
                        <figure class="image is-48x48">
                        <img src="https://bulma.io/assets/images/placeholders/96x96.png" alt="Placeholder image" />
                        </figure>
                    </div>
                    <div class="media-content">
                        <p class="title is-4"><a href="{{ .URL }}">{{ .Title }}</a></p>
                        <p class="subtitle is-6"><span class="tag">{{ .FeedName }}</span> {{ .CreatedAt | ToDateTime }}</p>
                    </div>
                    </div>
                    {{ if .Content }}
                    <p>{{ .Content }}</p>
                    {{ end }}
                {{ end }}
            {{ else }}
            <h3 class="title">{{ .Tr.Msg "no unread items" }}</h3>
            {{ end }}
        </div>
    </div>
{{ block "paging" .}}{{ end }}
</div>
{{ end }}
