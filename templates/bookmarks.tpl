{{ define "full-content" }}
<div class="bookmark-wrapper">
    {{ if not .Bookmarks }}
    <h3 class="title">No public bookmarks found</h3>
    {{ else }}
    <div class="is-pulled-right">
        <a href="{{ AddURLParam .URL "format=rss" }}">RSS<span class="icon"><i class="fas fa-rss"></i></span></a>
    </div>
    <h3 class="title">Public bookmarks ({{ .BookmarkCount }})</h3>
    {{ end }}
    <div class="columns">
        <div class="column is-2-fullhd is-one-quarter-desktop is-one-third-tablet bookmark-sidebar">
            <div class="content">
                <form action="" method="get">
                            {{ block "textFilter" .}}{{ end }}
                            <span class="select">
                                <select name="order_by">
                                    <option value="date_desc"{{ if eq .OrderBy "date_desc" }} selected="selected"{{ end }}>Date desc</option>
                                    <option value="date_asc"{{ if eq .OrderBy "date_asc" }} selected="selected"{{ end }}>Date asc</option>
                                </select>
                            </span>
                            <input type="submit" value="sort" class="button" />
                            <details {{ if .HasSearch }}open{{ end }}>
                                <summary>Advanced search options</summary>
                                    <div class="my-bookmarks__advanced-content">
                                        <div class="my-bookmarks__advanced-search">
                                                {{ block "domainFilter" .}}{{ end }}
                                                {{ block "ownerFilter" .}}{{ end }}
                                                {{ block "tagFilter" .}}{{ end }}
                                                {{ block "dateFilter" .}}{{ end }}
                                                {{ block "searchParameters" .}}{{ end }}
                                        </div>
                                        {{ block "submit" "Search" }}{{ end }}
                                    </div>
                            </details>
                </form>
                <div class="mt-5 is-hidden-mobile">
                    <h3>Frequent Tags</h3>
                    <div class="field is-grouped is-grouped-multiline ">
                        {{ range .FrequentTags }}
                        <div class="control">
                            <a class="tags has-addons" href="{{ URLFor "Public bookmarks" }}?tag={{ .Tag }}">
                                <span class="tag is-primary is-medium">{{ .Tag }}</span>
                                <span class="tag is-dark is-medium">{{ .Count }}</span>
                            </a>
                        </div>
                        {{ end }}
                    </div>
                </div>
            </div>
        </div>
        {{ $uid := 0 }}
        {{ if .User }}
        {{ $uid = .User.ID }}
        {{ end }}
        {{ $page := .Page }}
        {{ $csrf := .CSRF }}
        <div class="column bookmark-list">
            {{ range .Bookmarks }}
                {{ block "bookmark" KVData "Bookmark" . "UID" $uid "Page" $page "CSRF" $csrf }}{{ end }}
            {{ end }}
        </div>
    </div>
{{ block "paging" .}}{{ end }}
</div>
{{ end }}
