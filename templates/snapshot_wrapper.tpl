{{ define "full-content" }}
<div class="snapshot__container">
    <div class="container content my-5">
        <h3 class="title mb-0">Snapshot of <a href="{{ .Bookmark.URL }}">{{ Truncate .Bookmark.URL 100 }}</a></h3>
        <p>
            <strong>{{ .Snapshot.CreatedAt | ToDate }}</strong>
            <span class="tag is-info is-light">{{ .Snapshot.Size | FormatSize }}</span> <a href="{{ SnapshotURL .Snapshot.Key }}"><small>Fullscreen</small></a>
            - <a href="{{ URLFor "Download snapshot" }}?sid={{ .Snapshot.Key }}"><small>Download</small></a>
            - <a href="{{ URLFor "Snapshot details" }}?sid={{ .Snapshot.Key }}"><small>Details</small></a>
        </p>
        {{ if .OtherSnapshots }}
        <details>
            <summary>Other snapshots of this URL</summary>
            <div>
                {{ $os := .Snapshot }}
                {{ range $i,$s := .OtherSnapshots }}
                <div class="level"><div class="level-left">
                    <a href="{{ URLFor "Snapshot" }}?sid={{ $s.Key }}&bid={{ $s.Bid }}">{{ if $s.Title }}{{ $s.Title }}{{ else }}#{{ $i }} - {{ $s.CreatedAt | ToDate }}{{ end }}</a>
                    <a href="{{ URLFor "Snapshot diff" }}?s1={{ $os.Key }}&s2={{ $s.Key }}" class="button is-small">Compare</a>
                </div></div>
                {{ end }}
            </div>
        </details>
        {{ end }}
    </div>
</div>
<iframe src="{{ SnapshotURL .Snapshot.Key }}" title="snapshot of {{ .Bookmark.URL }}" class="snapshot-iframe"></iframe>
{{ end }}
