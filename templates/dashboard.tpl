{{ define "content" }}
<div class="content">
    <h4 class="title">Statistics</h4>
    <nav class="level">
        <div class="level-item has-text-centered">
            <div>
                <p class="heading">New bookmarks last week</p>
                <p class="title is-size-1">{{ .WeeklyBookmarkCount }}</p>
            </div>
        </div>
        <div class="level-item has-text-centered">
            <div>
                <p class="heading">New bookmarks this month</p>
                <p class="title is-size-1">{{ .MonthlyBookmarkCount }}</p>
            </div>
        </div>
        <div class="level-item has-text-centered">
            <div>
                <p class="heading">New bookmarks this year</p>
                <p class="title is-size-1">{{ .YearlyBookmarkCount }}</p>
            </div>
        </div>
    </nav>
    {{ if .Tags }}
    <h4 class="title">My frequent tags</h4>
    <div class="field is-grouped is-grouped-multiline">
        {{ range .Tags }}
        <div class="control">
            <a class="tags has-addons" href="{{ URLFor "My bookmarks" }}?tag={{ .Tag }}">
                <span class="tag is-primary is-medium">{{ .Tag }}</span>
                <span class="tag is-dark is-medium">{{ .Count }}</span>
            </a>
        </div>
        {{ end }}
    </div>
    {{ else }}
    {{ block "info" "Add tags to your bookmarks to be able to efficiently filter" }}{{ end }}
    {{ end }}
    {{ if .Bookmarks }}
      <h4 class="title">My latest bookmarks</h4>
      {{ $uid := .User.ID }}
      {{ $page := .Page }}
      {{ $csrf := .CSRF }}
      {{ $Tr := .Tr }}
      {{ range .Bookmarks }}
        {{ block "bookmark" KVData "Bookmark" . "UID" $uid "Page" $page "CSRF" $csrf "Tr" $Tr }}{{ end }}
      {{ end }}
    {{ else }}
      {{ block "warning" KVData "Warning" "No bookmarks added yet" "Tr" .Tr }}{{ end }}
    {{ end }}
</div>
{{ end }}
