{{ define "full-content" }}
<div class="fullscreen-wrapper">
    <h3 class="title">{{ .Tr.Msg "feeds" }}</h3>
    <div class="columns">
        {{ block "feedSidebar" . }}{{ end }}
        {{ $Tr := .Tr }}
        <div class="column container">
            {{ if .UnreadItemCount }}
                <h3 class="title">{{ .Tr.Msg "unread items" }} ({{ .UnreadItemCount }})</h3>
                {{ range .UnreadItems }}
                    {{ if HasAttr . "FeedName" }}
                        {{ block "feedItem" KVData "Item" . "Tr" $Tr  }}{{ end }}
                    {{ else }}
                        {{ block "feedBookmarkItem" KVData "Bookmark" . "Tr" $Tr }}{{ end }}
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
