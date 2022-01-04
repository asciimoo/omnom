{{ define "content" }}
<div class="snapshot__container">
<section id="home" class="hero is-medium">
    <h3>Snapshot of {{ .Bookmark.URL }}</h3>
    <h4>{{ .Snapshot.CreatedAt }}</h4>
    <p><a href="{{ BaseURL "/view_snapshot" }}?id={{ .Snapshot.Key }}">Fullscreen snapshot</a></p>
</section>
<div class="iframe-box">
    <div class="iframe-container">
        <iframe src="{{ BaseURL "/view_snapshot" }}?id={{ .Snapshot.Key }}" title="snapshot of {{ .Bookmark.URL }}" width="100%" height="100%" frameborder="1px"></iframe>
    </div>
</div>
</div>
{{ end }}
