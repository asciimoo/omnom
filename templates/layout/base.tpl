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
        {{ if .User }}
          <a href="/my_bookmarks" class="navbar-item{{ if eq .Page "my-bookmarks" }} is-active{{ end }}">My bookmarks</a>
        {{ end }}
        <a href="/bookmarks" class="navbar-item{{ if eq .Page "bookmarks" }} is-active{{ end }}">Public bookmarks</a>
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
        {{ if .Error }}
        {{ block "error" .Error }}{{ end }}
        {{ end }}
        {{ if .Warning }}
        {{ block "warning" .Warning }}{{ end }}
        {{ end }}
        {{ if .Info }}
        {{ block "info" .Info }}{{ end }}
        {{ end }}

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

{{ define "error" }}
<article class="message is-danger">
  <div class="message-body">{{ . | ToHTML }}</div>
</article>
{{ end }}

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
  <div class="message-body">{{ . | ToHTML }}</div>
</article>
{{ end }}


{{ define "note" }}
<article class="message is-info">
  <div class="message-header">
    <p>Note</p>
  </div>
  <div class="message-body">{{ . | ToHTML }}</div>
</article>
{{ end }}

{{ define "bookmark" }}
<div class="box media">
    <div class="media-content">
    <h4 class="title">{{ if .Favicon }}<img src="{{ .Favicon | ToURL }}" alt="favicon" /> {{ end }}<a href="{{ .URL }}" target="_blank">{{ .Title }}</a></h4>
    <p>{{ .Notes }}</p>
    </div>
    <div class="media-right">
        {{ range $i,$s := .Snapshots }}
        <a href="/snapshot?id={{ $s.ID }}">snapshot #{{ $i }}</a>
        {{ end }}
        {{ .CreatedAt | ToDate }} {{ if .Public }}Public{{ else }}Private{{ end }}
    </div>
</div>
{{ end}}

{{ define "paging" }}
<div class="columns is-centered">
    <div class="column is-narrow">
        {{ if gt .Pageno 1 }}
        <a href="?pageno={{ dec .Pageno }}" class="button is-primary is-medium is-outlined"><span class="icon"><i class="fas fa-angle-left"></i></span><span>Previous page</span></a>
        {{ end }}
        {{ if .HasNextPage }}
        <a href="?pageno={{ inc .Pageno }}" class="button is-primary is-medium is-outlined"><span>Next page</span><span class="icon"><i class="fas fa-angle-right"></i></span></a>
        {{ end }}
    </div>
</div>
{{ end }}
