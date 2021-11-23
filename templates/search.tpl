{{ define "content" }}
<div class="content"><form action="/search" method="get">
    <h2 class="title">Search</h2>
    <div class="field is-horizontal">
        <div class="field-label is-normal">
            <label class="label">Query</label>
        </div>
        <div class="field-body">
            <div class="field">
                <div class="control">
                    <input class="input" type="text" placeholder="search text.." name="query" value="{{ .SearchParams.Q }}">
                </div>
            </div>
        </div>
    </div>

    <div class="field is-horizontal">
        <div class="field-label is-normal">
        </div>
        <div class="field-body">
            <div class="field is-grouped">
                <div class="control">
                    <label class="checkbox">
                        <input type="checkbox" value="true" name="search_in_snapshot"{{ if .SearchParams.SearchInSnapshot }} checked="checked"{{ end }}>
                        Search in snapshots
                    </label>
                </div>
                <div class="control">
                    <label class="checkbox">
                        <input type="checkbox" value="true" name="search_in_note"{{ if .SearchParams.SearchInNote }} checked="checked"{{ end }}>
                        Search in notes
                    </label>
                </div>
                <div class="control">
                    <label class="checkbox">
                        <input type="checkbox" value="true" name="public"{{ if .SearchParams.IsPublic }} checked="checked"{{ end }}>
                        Only public bookmarks
                    </label>
                </div>
            </div>
        </div>
    </div>

    <div class="field is-horizontal">
        <div class="field-label is-normal">
            <label class="label">Domain</label>
        </div>
        <div class="field-body">
            <div class="field">
                <div class="control">
                    <input class="input" type="text" placeholder="domain name.." name="domain" value="{{ .SearchParams.Domain }}">
                </div>
            </div>
        </div>
    </div>

    <div class="field is-horizontal">
        <div class="field-label is-normal">
            <label class="label">Owner</label>
        </div>
        <div class="field-body">
            <div class="field">
                <div class="control">
                    <input class="input" type="text" placeholder="username.." name="owner" value="{{ .SearchParams.Owner }}">
                </div>
            </div>
        </div>
    </div>

    <div class="field is-horizontal">
        <div class="field-label is-normal">
            <label class="label">Tag</label>
        </div>
        <div class="field-body">
            <div class="field">
                <div class="control">
                    <input class="input" type="text" placeholder="tag.." name="tag" value="{{ .SearchParams.Tag }}">
                </div>
            </div>
        </div>
    </div>

    <div class="field is-horizontal">
        <div class="field-label is-normal">
            <label class="label">Date from/to</label>
        </div>
        <div class="field-body">
            <div class="field">
                <p class="control is-expanded">
                    <input class="input" type="text" placeholder="YYYY.MM.DD" name="from" value="{{ .SearchParams.FromDate }}">
                </p>
            </div>
            <div class="field">
                <p class="control is-expanded">
                    <input class="input" type="email" placeholder="YYYY.MM.DD" name="to" value="{{ .SearchParams.ToDate }}">
                </p>
            </div>
        </div>
    </div>

    <div class="field is-horizontal">
        <div class="field-label">
            <!-- Left empty for spacing -->
        </div>
        <div class="field-body">
            <div class="field">
                <div class="control">
                    <input type="submit" name="submit" value="Search" class="button is-primary" />
                </div>
            </div>
        </div>
    </div>

    {{ if not .Bookmarks }}
    <h3 class="title">No bookmarks found</h3>
    {{ else }}
    <h3 class="title">Search results{{ if .SearchParams.Q }} for "{{ .SearchParams.Q }}"{{ end }} ({{ .BookmarkCount }})</h3>
      {{ range .Bookmarks }}
        {{ block "bookmark" .}}{{ end }}
      {{ end}}
    {{ end }}
    {{ block "paging" .}}{{ end }}
</form></div>
{{ end }}
