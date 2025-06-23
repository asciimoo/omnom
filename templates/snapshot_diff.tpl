{{ define "full-content" }}
<div class="fullscreen-wrapper content">
    <h2 class="title">Snapshot diff of {{ Truncate .SURL 250 }}</h2>
    <p>Compared snapshots:
        <ol>
            <li><a href="{{ URLFor "Snapshot" }}?sid={{ .S1.Key }}&bid={{ .S1.BookmarkID }}">{{ .S1.CreatedAt | ToDate }}</a></li>
            <li><a href="{{ URLFor "Snapshot" }}?sid={{ .S2.Key }}&bid={{ .S2.BookmarkID }}">{{ .S2.CreatedAt | ToDate }}</a></li>
        </ol>
    <a href="{{ URLFor "Snapshot diff side by side" }}?s1={{ .S1.Key }}&s2={{ .S2.Key }}">Compare side by side</a>
    </p>
    <div class="columns">
        <div class="column">
        {{ if .LinkDiffs }}
            <h3>Link changes ({{ len .LinkDiffs }})</h3>
            <p>
                {{ range .LinkDiffs }}
                <div class="message {{ if eq .Type "+" }}is-primary{{ else}}is-danger{{ end }}">
                    <div class="message-body">
                        {{ if .Link.Text }}
                        <b>{{ .Link.Text }}</b><br />
                        {{ end }}
                        <a href="{{ .Link.Href }}">{{ .Link.Href }}</a>
                    </div>
                </div>
                {{ end }}
            </p>
        {{ else }}
            <h3>No Link changes</h3>
        {{ end }}
        </div>
        <div class="column">
        {{ if .TextDiffLen }}
            <h3>Text changes ({{ .TextDiffLen }})</h3>
            <p>
            {{ .TextDiff }}
            </p>
        {{ else }}
            <h3>No text changes</h3>
        {{ end }}
        </div>
        <div class="column">
        {{ if .ImageDiffs }}
            <h3>Multimedia changes ({{ len .ImageDiffs }})</h3>
            <p>
                {{ range .ImageDiffs }}
                <div class="content pl-4">
                    <figure class="imgdiff has-text-centered has-background-{{ if eq .Type "+" }}success{{ else }}danger-light{{ end }}">
                        <img src="{{ .Text | ResourceURL }}" class="has-ratio" />
                    </figure>
                </div>
                {{ end }}
            </p>
        {{ else }}
            <h3>No multimedia changes</h3>
        {{ end }}
        </div>
    </div>
</div>
{{ end }}
