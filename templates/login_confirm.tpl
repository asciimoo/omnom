{{ define "content" }}
<div class="columns">
<div class="column is-half is-offset-one-quarter">
<div class="card"><div class="card-content">
  <h3 class="title">{{ .Tr.Msg "successful login" }}<br />{{ .Tr.Msg "check your verification email" }}</h3>
</div></div>
</div>
</div>
{{ end }}
