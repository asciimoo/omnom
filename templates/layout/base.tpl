<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8" />
    <title>Omnom</title>    
    <link rel="stylesheet" href="{{ BaseURL "/static/css/style.css" }}" />
    <link rel="icon" type="image/png" href="{{ BaseURL "/static/ext/icons/omnom128.png" }}" sizes="128x128">

    {{ block "head" . }} {{ end }}
</head>
<body id="omnom-webapp">
<nav class="navbar shadow-bottom {{ block "content-class" . }}{{ end }}" role="navigation" aria-label="main navigation">
  <div class="navbar__container">
    <div class="navbar-brand is-size-4">
      <a class="navbar__logo" href="{{ BaseURL "/" }}"><span>om</span><span class="text--primary">nom</span> </a>
      <label for="nav-toggle-state" role="button" class="navbar-burger burger has-text-black" aria-label="menu" aria-expanded="false">
        <span aria-hidden="true"></span>
        <span aria-hidden="true"></span>
        <span aria-hidden="true"></span>
      </label>
    </div>
    <input type="checkbox" id="nav-toggle-state" />

    <div id="navbar-menu" class="navbar-menu is-size-5">
      <div class="navbar-start">
        <a href="{{ BaseURL "/" }}" class="navbar-item{{ if or (eq .Page "index") (eq .Page "dashboard") }} is-active{{ end }}"  >Our mission</a>
        {{ if .User }}
          <a href="{{ BaseURL "/my_bookmarks" }}" class="navbar-item{{ if eq .Page "my-bookmarks" }} is-active{{ end }}">My bookmarks</a>
        {{ end }}
        <a href="{{ BaseURL "/bookmarks" }}" class="navbar-item{{ if eq .Page "bookmarks" }} is-active{{ end }}">Public bookmarks</a>
      </div>
      <div class="navbar-end">
        {{ if .User }}
            <a href="{{ BaseURL "/profile" }}" class="navbar-item"><i class="fas fa-user"></i> &nbsp; {{ .User.Username }}</a>
            <div class="navbar-item"><a href="{{ BaseURL "/logout" }}" class="button is-outlined is-info">Logout</a></div>
        {{ else }}
            <div class="navbar-item"><a href="{{ BaseURL "/login" }}" class="button is-outlined is-info">Login</a></div>
            <div class="navbar-item"><a href="{{ BaseURL "/signup" }}" class="button is-outlined is-info">Signup</a></div>
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
        {{ block "extra_content" . }}{{ end }}
    </div>
</div>
{{ if (not .hideFooter) }}
<footer class="footer">
  <div class="container">
    <div class="content has-text-centered">
      <p>
          <strong>Omnom</strong> © 2021
      </p>
    </div>
  </div>
</footer>
{{ end }}
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
        <h4 class="title">
            <span class="icon-text">
                {{ if .Bookmark.Favicon }}
                <span class="icon">
                    <img src="{{ .Bookmark.Favicon | ToURL }}" alt="favicon" />
                </span>
                {{ end }}
                <span>
                    <a href="{{ .Bookmark.URL }}" target="_blank">{{ .Bookmark.Title }}</a>
                </span>
            </span>
            <p class="is-size-7 has-text-grey has-text-weight-normal">{{ Truncate .Bookmark.URL 100 }}</p>
        </h4>
        <p>{{ .Bookmark.Notes }}</p>
        {{ if .Bookmark.Tags }}
          {{ range .Bookmark.Tags }}
            <a href="{{ if or (eq $.Page "bookmarks") (ne $.Bookmark.UserID $.UID) }}{{ BaseURL "/bookmarks" }}{{ else }}{{ BaseURL "/my_bookmarks" }}{{ end }}?tag={{ .Text }}"><span class="tag is-info">{{ .Text }}</span></a>
          {{ end }}
        {{ end }}
    </div>
    <div class="media-right">
        {{ block "snapshots" .Bookmark.Snapshots }}{{ end }}
        {{ .Bookmark.CreatedAt | ToDate }} {{ if .Bookmark.Public }}Public{{ else }}Private{{ end }}
        {{ if eq .UID .Bookmark.UserID }}
          <a href="{{ BaseURL "/edit_bookmark" }}?id={{ .Bookmark.ID }}">edit</a>
        {{ else }}
          <a href="{{ BaseURL "/bookmark" }}?id={{ .Bookmark.ID }}">view</a>
        {{ end }}
    </div>
</div>
{{ end}}

{{ define "snapshots" }}
    {{ range $i,$s := . }}
    <a href="{{ BaseURL "/snapshot" }}?sid={{ $s.Key }}&bid={{ $s.BookmarkID }}">snapshot #{{ $i }} {{ $s.Title }} {{ $s.CreatedAt | ToDate }}</a>
    {{ end }}
{{ end }}

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

{{ define "textFilter" }}
<div class="field is-horizontal">
    <div class="field-body">
        <div class="field">
            <div class="control has-icons-left has-icons-right">
                <input class="input" type="text" placeholder="Search" name="query" value="{{ .SearchParams.Q }}">
                 <span class="icon is-small is-left">
                <i class="fas fa-search"></i>
                </span>
                <span class="icon is-small is-right">
                <i class="fas fa-times-circle"></i>
                </span>
            </div>
        </div>
    </div>
</div>
{{end}}
{{define "searchParameters"}}        
        <div class="field field-row">
            <label class="label">Search in snapshots</label>
            <input class="switch is-rounded" value="true" type="checkbox" id="search_in_snapshot"  name="search_in_snapshot"{{ if .SearchParams.SearchInSnapshot }} checked="checked"{{ end }}>
            <label for="search_in_snapshot"></label>    
        </div>
        <div class="field field-row">
            <label class="label">Search in notes</label>
            <input class="switch is-rounded" value="true" type="checkbox" id="search_in_note" name="search_in_note"{{ if .SearchParams.SearchInNote }} checked="checked"{{ end }}>
            <label for="search_in_note"></label>    
        </div>
        {{ if eq .Page "my-bookmarks" }}
        <div class="field field-row">
            <label class="label">Only public bookmarks</label>
            <input class="switch is-rounded" value="true" type="checkbox" id="public" name="public"{{ if .SearchParams.IsPublic }} checked="checked"{{ end }}>            
            <label for="public"></label>    
        </div>
        {{ end }}    
{{ end }}

{{ define "domainFilter" }}
<div class="field">    
    <label class="label">Domain</label>
    <div class="control">
        <input class="input" type="text" placeholder="Insert Url" name="domain" value="{{ .SearchParams.Domain }}">
    </div>
</div>
{{ end }}

{{ define "ownerFilter" }}
<div class="field">
<label class="label">Owner</label>
    <div class="control">
        <input class="input" type="text" placeholder="username.." name="owner" value="{{ .SearchParams.Owner }}">
    </div>
</div>
{{ end }}

{{ define "tagFilter" }}
<div class="field">
<label class="label">Tags</label>
    <div class="control">
        <input class="input" type="text" placeholder="Add tag" name="tag" value="{{ .SearchParams.Tag }}">
    </div>
</div>
{{ end }}

{{ define "dateFilter" }}
<div class="field is-horizontal">
    <div class="field-body">
    <div class="field">
        <label class="label">Date from</label>
            <p class="control is-expanded">
                <input class="input" type="date" placeholder="YYYY.MM.DD" name="from" value="{{ .SearchParams.FromDate }}">
            </p>
        </div>
        <div class="field">
        <label class="label">Date to</label>
            <p class="control is-expanded">
                <input class="input" type="date" placeholder="YYYY.MM.DD" name="to" value="{{ .SearchParams.ToDate }}">
            </p>
        </div>
    </div>    
</div>
{{ end }}

{{ define "submit" }}
<div class="omnom-popup__submit">
    <input type="submit" name="submit" value="Search" class="button is-primary" />
</div>

{{ end }}
