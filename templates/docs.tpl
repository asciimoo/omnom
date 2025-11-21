{{ define "full-content" }}
<div class="fullscreen-wrapper">
    <h3 class="title">Documentation</h3>
    <div class="columns">
        <div class="column is-2-fullhd is-one-quarter-desktop is-one-third-tablet">
            <div class="mx-4">
                <ul>
                {{ $Doc := .Doc }}
                {{ range .TOC }}
                    <li><a class="is-size-4{{ if eq $Doc.Name .Name }} has-text-primary{{ end }}" href="{{ URLFor "docs" .Name }}">{{ .Title }}</a></li>
                    {{ $TOC := . }}
                    {{ if .Headings }}{{ range $_, $v := .Headings }}
                    <ul class="mx-4">
                        <li><a class="is-size-6" href="{{ URLFor "docs" $TOC.Name }}#{{ index $v 0 }}">{{ index $v 1 }}</a></li>
                    </ul>
                    {{ end }}{{ end }}
                {{ end }}
                </ul>
            </div>
        </div>
        <div class="column content container mx-0">
            {{ .Doc.Content }}
        </div>
    </div>
</div>
{{ end }}
