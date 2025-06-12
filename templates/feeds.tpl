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
        {{ $Tr := .Tr }}
        <div class="column">
            {{ if .UnreadItemCount }}
                <h3 class="title">{{ .Tr.Msg "unread items" }} ({{ .UnreadItemCount }})</h3>
                {{ range .UnreadItems }}
                    {{ if .FeedName }}
                        {{ block "unreadFeedItem" . }}{{ end }}
                    {{ else }}
                        {{ block "unreadBookmark" KVData "Bookmark" . "Tr" $Tr }}{{ end }}
                    {{ end }}
                {{ end }}
                <a href="#" class="button is-primary is-medium">Mark items as read</a>
            {{ else }}
            <h3 class="title">{{ .Tr.Msg "no unread items" }}</h3>
            {{ end }}
        </div>
    </div>
{{ block "paging" .}}{{ end }}
</div>
{{ end }}
