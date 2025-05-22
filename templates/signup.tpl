{{ define "content" }}
<form method="post" action="{{ URLFor "Signup" }}">
<div class="columns">
<div class="column is-half is-offset-one-quarter">
<div class="card"><div class="card-content">
  {{ if .OAuthID }}
    <h3 class="title">Successful authentication</h3>
    <h3>Assign username and e-mail address to your account</h3>
  {{ else }}
    <h4 class="title">Sign up</h4>
  {{ end }}
  <div class="content">
    <div class="field">
      <label class="label">Username</label>
      <div class="control has-icons-left">
        <input class="input" type="text" name="username" placeholder="username.."{{ if .OAuthUsername }} value="{{ .OAuthUsername }}"{{ end }} />
        <span class="icon is-small is-left">
          <i class="fas fa-user"></i>
        </span>
      </div>
    </div>
    <div class="field">
      <label class="label">Email</label>
      <div class="control has-icons-left">
        <input class="input" type="email" name="email" placeholder="email.."{{ if .OAuthEmail }} value="{{ .OAuthEmail }}"{{ end }} />
        <span class="icon is-small is-left">
          <i class="fas fa-envelope"></i>
        </span>
      </div>
    </div>
    <div class="field">
      <input type="submit" value="submit" class="button" />
    </div>
  </div>
</div></div>
{{ if (and .OAuth (not .OAuthID))}}
<div class="card"><div class="card-content has-text-centered">
    <h3 class="title">or sign up with</h3>
    {{ range $name, $attrs := .OAuth }}
    <a href="{{ URLFor "Oauth" }}?provider={{ $name }}"><i class="{{ $attrs.Icon }} fa-6x px-4"></i>
    {{ end }}
</div></div>
{{ end }}
</div>
</div>
</form>
{{ end }}
