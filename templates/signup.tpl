{{ define "content" }}
<form method="post" action="{{ BaseURL "/signup" }}">
<div class="columns">
<div class="column is-half is-offset-one-quarter">
<div class="card"><div class="card-content">
  <h3 class="title">Signup</h3>
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
      <label class="label">Email</label>
      <div class="control has-icons-left">
        <input class="input" type="email" name="email" placeholder="email..">
        <span class="icon is-small is-left">
          <i class="fas fa-envelope"></i>
        </span>
      </div>
    </div>
    <div class="field">
      <input type="submit" value="signup" class="button" />
    </div>
  </div>
</div></div>
</div>
</div>
</form>
{{ end }}
