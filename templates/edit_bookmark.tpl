{{ define "full-content" }}
<div class="fullscreen-wrapper">
    <div class="columns">
        <div class="column is-offset-2">
            <h2 class="title">Edit bookmark <i>"{{ .Bookmark.Title }}"</i><br />
                <span class="icon-text">
                    {{ if .Bookmark.Favicon }}
                    <span class="icon">
                        <img src="{{ .Bookmark.Favicon | ToURL }}" alt="favicon" />
                    </span>
                    {{ end }}
                    <span class="is-size-7 has-text-grey has-text-weight-normal">{{ Truncate .Bookmark.URL 200 }}</span>
                </span>
            </h2>
        </div>
    </div>
    <div class="columns">
        <div class="column is-offset-2">
            <form method="post" action="{{ URLFor "Save bookmark" }}">
                <input type="hidden" name="id" value="{{ .Bookmark.ID }}" />
                <div class="field">
                    <label class="label">Title</label>
                    <div class="control">
                        <input class="input" type="text" name="title" value="{{ .Bookmark.Title }}" />
                    </div>
                </div>
                {{ block "collectionFilter" . }}{{ end }}
                <div class="field">
                    <div class="control">
                        <label class="checkbox">
                            <b>Public</b>
                            <input name="public" type="checkbox"{{ if .Bookmark.Public }} checked{{ end }}>
                        </label>
                    </div>
                </div>
                <div class="field">
                    <div class="control">
                        <label class="checkbox">
                            <b>Unread</b>
                            <input name="unread" type="checkbox"{{ if .Bookmark.Unread }} checked{{ end }}>
                        </label>
                    </div>
                </div>
                <div class="field">
                    <label class="label">Notes</label>
                    <div class="control">
                        <textarea class="textarea" name="notes">{{ .Bookmark.Notes }}</textarea>
                    </div>
                </div>
                <div class="field">
                    <div class="control">
                        <input class="button is-primary" type="submit" value="Save" />
                    </div>
                </div>
            </form>
            <div class="field is-grouped is-grouped-right">
                <form method="post" action="{{ URLFor "Delete bookmark" }}">
                    <input class="button is-danger" type="submit" value="Delete this bookmark" />
                    <input type="hidden" name="id" value="{{ .Bookmark.ID }}" />
                </form>
            </div>
        </div>
        <div class="column is-offset-1">
            <h3 class="title is-size-4">Tags</h3>
            <div class="bookmark__tags mb-4">
                {{ if .Bookmark.Tags }}
                {{ range .Bookmark.Tags }}
                    <span class="tag is-muted-primary is-large">{{ .Text }}
                        <form method="post" action="{{ URLFor "Delete tag" }}">
                            <input type="hidden" name="tid" value="{{ .ID }}"/>
                            <input type="hidden" name="bid" value="{{ $.Bookmark.ID}}"/>
                            <button class="delete is-large ml-3" type="submit"></button>
                        </form>
                    </span>
                {{ end }}
                {{ end }}
            </div>
            <form method="post" action="{{ URLFor "Add tag" }}">
                <label class="label">Add tag</label>
                <div class="field has-addons">
                    <div class="control">
                        <input class="input" type="text" name="tag" />
                    </div>
                    <div class="control">
                        <input class="button is-primary" type="submit" value="Add" />
                        <input type="hidden" name="bid" value="{{ .Bookmark.ID }}" />
                    </div>
                </div>
            </form>

            {{ if .Bookmark.Snapshots }}
            <h3 class="title is-size-4 mt-6">Snapshots</h3>
            <div class="columns is-mobile">
                <div class="column is-narrow">
                    <div class="list">
                        <dl>
                            {{ range .Bookmark.Snapshots }}
                            <div class="list-item">
                                <li class="pure-list">
                                    <form method="post" action="{{ URLFor "Delete snapshot" }}">
                                        <h4 class="has-text-dark"><a href="{{ BaseURL "/snapshot" }}?sid={{ .Key }}&bid={{ $.Bookmark.ID }}">{{ .Title }} {{ .CreatedAt | ToDate }}</a>
                                            <input type="hidden" name="bid" value="{{ $.Bookmark.ID }}" />
                                            <input type="hidden" name="sid" value="{{ .ID }}" />
                                            <input type="submit" class="button is-danger is-small" value="Delete" />
                                        </h4>
                                    </form>
                                </li>
                            </div>
                            {{ end}}
                        </dl>
                    </div>
                </div>
            </div>
            {{ end }}
        </div>
    </div>
</div>
{{ end }}
