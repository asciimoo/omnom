{{ define "full-content" }}
<div class="snapshot__container">
    <div class="container content my-5">
    {{ if .Snapshot }}
        <h3 class="title mb-0">Snapshot of <a href="{{ .URL }}">{{ Truncate .URL 100 }}</a></h3>
        <p>
            <strong>{{ .Snapshot.CreatedAt | ToDate }}</strong>
            <span class="tag is-info is-light">{{ .Snapshot.Size | FormatSize }}</span> <a href="{{ SnapshotURL .Snapshot.Key }}"><small>Fullscreen</small></a>
            - <a href="{{ URLFor "Download snapshot" }}?sid={{ .Snapshot.Key }}"><small>Download</small></a>
            - <a href="{{ URLFor "Snapshot details" }}?sid={{ .Snapshot.Key }}"><small>Details</small></a>
        </p>
    {{ else }}
        <h3>No snapshot found</h3>
        {{ if .AllowSnapshotCreation }}
        <a href="{{ URLFor "Create bookmark form" }}?url={{ .URL }}">Create snapshot</a>
        {{ else }}
        <h4>Server side snapshot creation is disabled</h4>
        <p>Create a snapshot with your extension by visiting <a href="{{ .URL }}">{{ Truncate .URL 100 }}</a></p>
        {{ end }}
    {{ end }}
    </div>
</div>
{{ if .Snapshot }}
<iframe src="{{ SnapshotURL .Snapshot.Key }}" title="snapshot of {{ .URL }}" class="snapshot-iframe"></iframe>
{{ end }}
{{ end }}
