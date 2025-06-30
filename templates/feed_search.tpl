{{ define "full-content" }}
<div class="fullscreen-wrapper">
    <h3 class="title">{{ .Tr.Msg "feed search" }}</h3>
    <div class="columns">
        {{ block "feedSidebar" . }}{{ end }}
        {{ $Tr := .Tr }}
        <div class="column">
            {{ if .ItemCount }}
                {{ if .FeedName }}
                <h3 class="title">{{ .FeedName }} ({{ .ItemCount }})</h3>
                {{ else }}
                <h3 class="title">{{ .Tr.Msg "search results" }} ({{ .ItemCount }})</h3>
                {{ end }}
                {{ range .Items }}
                    {{ if HasAttr . "FeedName" }}
                        {{ block "feedItem" KVData "Item" . "Tr" $Tr  }}{{ end }}
                    {{ else }}
                        {{ block "feedBookmarkItem" KVData "Bookmark" . "Tr" $Tr }}{{ end }}
                    {{ end }}
                {{ end }}
                {{/* TODO PAGINATION */}}
            {{ else }}
                <h3 class="title">{{ .Tr.Msg "no results found" }}</h3>
            {{ end }}
        </div>
    </div>
</div>
{{ end }}
