{{ define "content" }}
<form method="post" action="{{ URLFor "Login" }}">
<div class="columns">
<div class="column is-half is-offset-one-quarter">
<div class="card"><div class="card-content">
  <h3 class="title">Login</h3>
  <div class="content">
    <div class="field">
      <label class="label">Username</label>
      <div class="control has-icons-left">
        <input class="input {{ if .Error }}is-danger{{ end }}" type="text" name="username" placeholder="username.." />
        <span class="icon is-small is-left">
          <i class="fas fa-user"></i>
        </span>
      </div>
      {{ if .Error }}
          <p class="help is-danger">{{ .Error }}</p>
      {{ end }}
    </div>
    <div class="field">
      <input type="submit" value="login" class="button" />
    </div>
  </div>
</div></div>
</div>
</div>
</form>
{{ end }}
