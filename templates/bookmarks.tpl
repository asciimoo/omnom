{{ define "content" }}
<div class="content">
    {{ if not .Bookmarks }}
    <h3 class="title">No public bookmarks found</h3>
    {{ else }}
    <h3 class="title">Public bookmarks ({{ .BookmarkCount }})</h3>
    {{ end }}
    <div class="content">
    <form action="" method="get">
    {{ block "textFilter" .}}{{ end }}
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
    {{ $uid := 0 }}
    {{ if .User }}
      {{ $uid = .User.ID }}
    {{ end }}
    {{ $page := .Page }}
    {{ range .Bookmarks }}
        {{ block "bookmark" KVData "Bookmark" . "UID" $uid "Page" $page }}{{ end }}
    {{ end }}
{{ block "paging" .}}{{ end }}
</div>
{{ end }}
