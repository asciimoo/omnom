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
                {{ $Tr := .Tr }}
                {{ range .Feeds }}
                <h4>
                    <div class="is-pulled-right">
                        <a href="{{ URLFor "edit feed" }}?id={{ .ID }}" aria-label="{{ $Tr.Msg "edit feed" }}"><span class="icon"><i class="fas fa-pencil"></i></span></a>
                    </div>
                    {{ .Name }}{{ if .UnreadCount }} <span class="tag is-medium">{{ .UnreadCount }}</span>{{ end }}
                </h4>
                {{ end }}
            </div>
        </div>
        {{ $uid := 0 }}
        {{ $Tr := .Tr }}
        <div class="column">
            {{ if .UnreadItemCount }}
                <h3 class="title">{{ .Tr.Msg "unread items" }} ({{ .UnreadItemCount }})</h3>
                {{ range .UnreadItems }}
                    {{ if HasAttr . "FeedName" }}
                        {{ block "unreadFeedItem" KVData "Item" . "Tr" $Tr  }}{{ end }}
                    {{ else }}
                        {{ block "unreadBookmark" KVData "Bookmark" . "Tr" $Tr }}{{ end }}
                    {{ end }}
                {{ end }}
                <div class="columns is-centered">
                    <div class="column is-narrow">
                        <form method="post" action="{{ URLFor "archive items" }}">
                            <input type="hidden" name="bids" value="{{ .BookmarkIDs }}" />
                            <input type="hidden" name="fids" value="{{ .FeedItemIDs }}" />
                            <input type="submit" class="button is-primary is-medium" value="{{ .Tr.Msg "archive page" }}" />
                        </form>
                    </div>
                </div>
            {{ else }}
            <h3 class="title">{{ .Tr.Msg "no unread items" }}</h3>
            {{ end }}
        </div>
    </div>
</div>
{{ end }}
