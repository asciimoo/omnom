{{ template "base.html.tpl" . }}

{{ define "content" }}
You can log in <a href="{{ .BaseURL }}?token={{ .Token }}">here</a>.
{{ end }}
