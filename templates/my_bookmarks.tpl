{{ define "content" }}
<div class="content">
    {{ if not .Bookmarks }}
    <h3 class="title">No bookmarks found</h3>
    {{ else }}
    <div class="is-pulled-right"><a href="{{ AddURLParam .URL "format=rss" }}">RSS<span class="icon"><i class="fas fa-rss"></i></span></a></div>
    <h3 class="title">My bookmarks ({{ .BookmarkCount }})</h3>
    {{ end }}
    <div class="content">
    <form action="" method="get">
    <div class="columns mb-2">
        <div class="column">
        {{ block "textFilter" .}}{{ end }}
        </div>
        <div class="column is-narrow has-text-right">
            <span class="select">
                <select name="order_by">
                    <option value="date_desc"{{ if eq .OrderBy "date_desc" }} selected="selected"{{ end }}>Date desc</option>
                    <option value="date_asc"{{ if eq .OrderBy "date_asc" }} selected="selected"{{ end }}>Date asc</option>
                </select>
            </span>
            <input type="submit" value="sort" class="button" />
        </div>
    </div>
    <div class="accordion-tabs">
        <div class="accordion-tab">
          <input {{ if .HasSearch }} checked{{ end }} class="accordion-tab__control" type="checkbox" id="chck2">
          <label class="accordion-tab-label" for="chck2">
            <div class="my-bookmarks__section-header">
              <h3>
                Advanced search options
              </h3>
              <i class="fas fa-angle-down"></i>
            </div>
          </label>
          <div class="accordion-tab-content">
          <div class="my-bookmarks__advanced-content">
            <div class="my-bookmarks__advanced-search">
                <div class="my-bookmarks__search-params">
                <p class="my-bookmarks__h3">Search parameters</p>
                    {{ block "searchParameters" .}}{{ end }}
                </div>
                <div class="my-bookmarks__query">
                <p class="my-bookmarks__h3">Query</p>
                {{ block "domainFilter" .}}{{ end }}
                {{ block "tagFilter" .}}{{ end }}
                {{ block "dateFilter" .}}{{ end }}
                </div>
            </div>
                {{ block "submit" . }}{{ end }}
          </div>
          </div>
        </div>
      </div>
    </form>
    </div>
    {{ $uid := .User.ID }}
    {{ $page := .Page }}
    {{ $csrf := .CSRF }}
    {{ range .Bookmarks }}
        {{ block "bookmark" KVData "Bookmark" . "UID" $uid "Page" $page "CSRF" $csrf }}{{ end }}
    {{ end }}
    {{ block "paging" .}}{{ end }}
    </div>
    {{ end }}
</div>
