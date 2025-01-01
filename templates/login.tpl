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
        <input class="input" type="text" name="username" placeholder="username.." />
        <span class="icon is-small is-left">
          <i class="fas fa-user"></i>
        </span>
      </div>
    </div>
    <div class="field">
      <input type="submit" value="submit" class="button" />
    </div>
  </div>
</div></div>
{{ if .OAuth }}
<div class="card"><div class="card-content has-text-centered">
    <h3 class="title">or sign in with</h3>
    {{ range $name, $attrs := .OAuth }}
    <a href="{{ URLFor "Oauth" }}?provider={{ $name }}"><i class="{{ $attrs.Icon }} fa-6x px-4"></i>
    {{ end }}
</div></div>
{{ end }}
</div>
</div>
</form>
{{ end }}
