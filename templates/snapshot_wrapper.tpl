{{ define "content" }}
<section id="home" class="hero is-medium">
    <h3>Snapshot of {{ .Bookmark.URL }}</h3>
    <h4>{{ .Snapshot.CreatedAt }}</h4>
    <p><a href="/viewSnapshot?id={{ .Snapshot.ID }}">Fullscreen snapshot</a></p>
</section>
{{ end }}
{{ define "extra_content" }}
<div class="iframe-box">
    <div class="iframe-container">
        <iframe src="/viewSnapshot?id={{ .Snapshot.ID }}" title="snapshot of {{ .Bookmark.URL }}" width="100%" height="100%" frameborder="1px"></iframe>
    </div>
</div>
{{ end }}
