{{ define "full-content" }}
<div class="fullscreen-wrapper">
    <h3 class="title">{{ .Tr.Msg "search" }}</h3>
    <div class="columns">
        <div class="column is-2-fullhd is-one-quarter-desktop is-one-third-tablet">
            <div class="content">
            </div>
        </div>
        {{ $Tr := .Tr }}
        <div class="column">
            {{ if .ItemCount }}
                <h3 class="title">{{ .Tr.Msg "search results" }} ({{ .ItemCount }})</h3>
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
