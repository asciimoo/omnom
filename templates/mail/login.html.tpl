{{ template "base.html.tpl" . }}

{{ define "content" }}
Hello {{ .Username }},<br />
<p>
You can log in to Omnom <a href="{{ .URL }}">here</a>.
</p>
Happy Omnoming
{{ end }}
