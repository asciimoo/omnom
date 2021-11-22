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
    {{ if .Bookmarks }}
      <h4 class="title">My latest bookmarks</h4>
      {{ range .Bookmarks }}
          {{ block "bookmark" .}}{{ end }}
      {{ end }}
    {{ else }}
      {{ block "warning" "No bookmarks added yet"}}{{ end }}
    {{ end }}
</div>
{{ end }}
