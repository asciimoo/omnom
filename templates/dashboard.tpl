{{ define "content" }}
<div class="content">
    <h4 class="title">{{ .Tr.Msg "statistics" }}</h4>
    <nav class="level">
        <div class="level-item has-text-centered">
            <div>
                <p class="heading">{{ .Tr.Msg "new bookmarks week" }}</p>
                <p class="title is-size-1">{{ .WeeklyBookmarkCount }}</p>
            </div>
        </div>
        <div class="level-item has-text-centered">
            <div>
                <p class="heading">{{ .Tr.Msg "new bookmarks month" }}</p>
                <p class="title is-size-1">{{ .MonthlyBookmarkCount }}</p>
            </div>
        </div>
        <div class="level-item has-text-centered">
            <div>
                <p class="heading">{{ .Tr.Msg "new bookmarks year" }}</p>
                <p class="title is-size-1">{{ .YearlyBookmarkCount }}</p>
            </div>
        </div>
    </nav>
    {{ if .Tags }}
    <h4 class="title">{{ .Tr.Msg "my frequent tags" }}</h4>
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
    {{ block "info" KVData "Info" (.Tr.Msg "add tag tip") }}{{ end }}
    {{ end }}
    {{ if .Bookmarks }}
      <h4 class="title">{{ .Tr.Msg "my latest bookmarks" }}</h4>
      {{ $uid := .User.ID }}
      {{ $page := .Page }}
      {{ $Tr := .Tr }}
      {{ range .Bookmarks }}
        {{ block "bookmark" KVData "Bookmark" . "UID" $uid "Page" $page "Tr" $Tr }}{{ end }}
      {{ end }}
    {{ else }}
      {{ block "warning" KVData "Warning" (.Tr.Msg "no bookmark found") "Tr" .Tr }}{{ end }}
    {{ end }}
</div>
{{ end }}
