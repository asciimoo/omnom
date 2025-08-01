{{ define "full-content" }}
<div class="fullscreen-wrapper">
    {{ if not .Bookmarks }}
    <h3 class="title">{{ .Tr.Msg "no bookmarks found" }}</h3>
    {{ else }}
    <div class="is-pulled-right"><a href="{{ AddURLParam .URL "format=rss" }}">RSS<span class="icon"><i class="fas fa-rss"></i></span></a></div>
    <h3 class="title">{{ .Tr.Msg "my bookmarks" }} ({{ .BookmarkCount }})</h3>
    {{ end }}
    <div class="columns">
        <div class="column is-2-fullhd is-one-quarter-desktop is-one-third-tablet bookmark-sidebar">
            <div class="content">
                <form action="" method="get">
                    {{ block "textFilter" .}}{{ end }}
                    {{ block "sortBy" . }}{{ end }}
                    <details {{ if .HasSearch }}open{{ end }} class="mt-4">
                        <summary>{{ .Tr.Msg "advanced search options" }}</summary>
                        <div class="my-bookmarks__advanced-content">
                            <div class="my-bookmarks__advanced-search">
                                {{ block "domainFilter" .}}{{ end }}
                                {{ block "ownerFilter" .}}{{ end }}
                                {{ block "collectionFilter" .}}{{ end }}
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
            </div>
            <div class="content collections">
                <div class="is-pulled-right">
                    <a href="{{ URLFor "edit collection form" }}" aria-label="{{ .Tr.Msg "add collection" }}"><span class="icon"><i class="fas fa-plus"></i></span></a>
                </div>
                {{ if .Collections }}
                    <h3>{{ .Tr.Msg "collections" }}</h3>
                    {{ block "collections" KVData "Collections" .Collections "CurrentCollection" .CurrentCollection "Tr" .Tr }}{{ end }}
                {{ else }}
                    <h3>{{ .Tr.Msg "no collection" }}</h3>
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
</div>
