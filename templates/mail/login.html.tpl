{{ template "base.html.tpl" . }}

{{ define "content" }}
Hello {{ .Username }},<br />
<p>
You can log in to Omnom <a href="{{ .BaseURL }}?token={{ .Token }}">here</a>.
</p>
Happy Omnoming
{{ end }}
