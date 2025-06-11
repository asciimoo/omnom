{{ define "full-content" }}
<div class="fullscreen-wrapper">
    {{ if not .Bookmarks }}
    <h3 class="title">{{ .Tr.Msg "no bookmarks found" }}</h3>
    {{ else }}
    <div class="is-pulled-right">
        <a href="{{ AddURLParam .URL "format=rss" }}">RSS<span class="icon"><i class="fas fa-rss"></i></span></a>
    </div>
    <h3 class="title">{{ .Tr.Msg "public bookmarks" }} ({{ .BookmarkCount }})</h3>
    {{ end }}
    <div class="columns">
        <div class="column is-2-fullhd is-one-quarter-desktop is-one-third-tablet bookmark-sidebar">
            <div class="content">
                <form action="" method="get">
                    {{ block "textFilter" . }}{{ end }}
                    {{ block "sortBy" . }}{{ end }}
                    <details {{ if .HasSearch }}open{{ end }}>
                        <summary>{{ .Tr.Msg "advanced search options" }}</summary>
                        <div class="my-bookmarks__advanced-content">
                            <div class="my-bookmarks__advanced-search">
                                {{ block "domainFilter" .}}{{ end }}
                                {{ block "ownerFilter" .}}{{ end }}
                                {{ block "tagFilter" .}}{{ end }}
                                {{ block "dateFilter" .}}{{ end }}
                                {{ block "searchParameters" .}}{{ end }}
                            </div>
                            {{ block "submit" (.Tr.Msg "search") }}{{ end }}
                            <div class="mt-5">
                                <a href="{{ URLFor "Snapshots" }}">{{ .Tr.Msg "snapshot search" }}</a>
                            </div>
                        </div>
                    </details>
                </form>
                {{ if .FrequentTags }}
                <div class="mt-5 is-hidden-mobile">
                    <h3>{{ .Tr.Msg "frequent tags" }}</h3>
                    <div class="field is-grouped is-grouped-multiline ">
                        {{ range .FrequentTags }}
                        <div class="control">
                            <a class="tags has-addons" href="{{ URLFor "Public bookmarks" }}?tag={{ .Tag }}">
                                <span class="tag is-muted-primary is-medium">{{ .Tag }}</span>
                                <span class="tag is-grey is-medium">{{ .Count }}</span>
                            </a>
                        </div>
                        {{ end }}
                    </div>
                </div>
                {{ end }}
            </div>
        </div>
        {{ $uid := 0 }}
        {{ if .User }}
        {{ $uid = .User.ID }}
        {{ end }}
        {{ $page := .Page }}
        {{ $Tr := .Tr }}
        <div class="column bookmark-list">
            {{ range .Bookmarks }}
                {{ block "bookmark" KVData "Bookmark" . "UID" $uid "Page" $page "Tr" $Tr }}{{ end }}
            {{ end }}
        </div>
    </div>
{{ block "paging" .}}{{ end }}
</div>
{{ end }}
