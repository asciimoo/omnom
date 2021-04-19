<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8" />
    <title>Omnom</title>
    <link rel="stylesheet" href="/static/css/bulma.min.css" />
    <link rel="stylesheet" href="/static/css/fa.min.css" />
    <link rel="stylesheet" href="/static/css/style.css" />
    <link rel="icon" type="image/png" href="/static/ext/icons/omnom128.png" sizes="128x128">

    {{ block "head" . }} {{ end }}
</head>
<body>
<nav class="navbar border-bottom {{ block "content-class" . }}{{ end }}" role="navigation" aria-label="main navigation">
  <div class="container">
    <div class="navbar-brand is-size-4">
      <a class="navbar-item{{ if eq .Page "index" }} is-active{{ end }}" href="/"><strong>Omnom</strong> </a>
      <label for="nav-toggle-state" role="button" class="navbar-burger burger has-text-black" aria-label="menu" aria-expanded="false">
        <span aria-hidden="true"></span>
        <span aria-hidden="true"></span>
        <span aria-hidden="true"></span>
      </label>
    </div>
    <input type="checkbox" id="nav-toggle-state" />

    <div id="navbar-menu" class="navbar-menu is-size-5">
      <div class="navbar-start">
        <a href="/bookmarks" class="navbar-item{{ if eq .Page "bookmarks" }} is-active{{ end }}">Bookmarks</a>
      </div>
      <div class="navbar-end">
        <div class="navbar-item">
          <p class="control has-icons-left">
              <input class="input is-rounded" type="text" placeholder="search">
              <span class="icon is-small is-left"><i class="fas fa-search"></i></span>
          </p>
        </div>
        {{ if .User }}
            <a href="/profile" class="navbar-item"><i class="fas fa-user"></i> &nbsp; {{ .User.Username }}</a>
            <div class="navbar-item"><a href="/logout" class="button is-outlined is-info">Logout</a></div>
        {{ else }}
            <div class="navbar-item"><a href="/login" class="button is-outlined is-info">Login</a></div>
            <div class="navbar-item"><a href="/signup" class="button is-outlined is-info">Signup</a></div>
        {{ end }}
      </div>
    </div>
  </div>
</nav>

<div class="section {{ block "content-class" . }}{{ end }}">
    <div class="bd-main-container container">

        {{ block "content" . }}{{ end }}
    </div>
</div>
{{ block "extra_content" . }}{{ end }}
<footer class="footer">
  <div class="container">
    <div class="content has-text-centered">
      <p>
          <strong>Omnom</strong> © 2021
      </p>
    </div>
  </div>
</footer>
</body>
</html>

{{ define "warning" }}
<article class="message is-warning">
  <div class="message-header">
    <p>Warning</p>
  </div>
  <div class="message-body">{{ . | ToHTML }}</div>
</article>
{{ end }}

{{ define "info" }}
<article class="message is-info">
  <div class="message-header">
    <p>Note</p>
  </div>
  <div class="message-body">{{ . | ToHTML }}</div>
</article>
{{ end }}

{{ define "attention" }}
<article class="message is-warning">
  <div class="message-body">{{ . | ToHTML }}</div>
</article>
{{ end }}
