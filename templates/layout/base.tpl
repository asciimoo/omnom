<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8" />
    <title>Omnom</title>    
    <link rel="stylesheet" href="{{ BaseURL "/static/css/style.css" }}" />
    <link rel="icon" type="image/png" href="{{ BaseURL "/static/icons/omnom.png" }}" sizes="128x128">

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
        <a href="{{ BaseURL "/" }}" class="navbar-item{{ if or (eq .Page "index") (eq .Page "dashboard") }} is-active{{ end }}">Home</a>
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
    </div>
</div>
{{ if (not .hideFooter) }}
<footer class="footer">
  <div class="container">
    <div class="content has-text-centered">
      <p>
          <strong>Omnom</strong> © 2022
      </p>
      <p>
          <a href="{{ BaseURL "/api" }}">API</a>
          | <a href="https://github.com/asciimoo/omnom">GitHub</a>
          | <a href="https://addons.mozilla.org/en-US/firefox/addon/omnom/">Firefox extension</a>
          | <a href="https://chrome.google.com/webstore/detail/omnom/nhpakcgbfdhghjnilnbgofmaeecoojei">Chrome extension</a>
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
<div class="box media bookmark__container">
    <div class="bookmark__header">
      <div class="bookmark__title">
        <div class="bookmark__favicon">
          {{ if .Bookmark.Favicon }}
            <span class="icon">
              <img src="{{ .Bookmark.Favicon | ToURL }}" alt="favicon" />
            </span>
            {{ end }}
        </div>
        <a href="{{ .Bookmark.URL }}" target="_blank">
          <h4 class="title">
                {{ .Bookmark.Title }}
            <p class="is-size-7 has-text-grey has-text-weight-normal">{{ Truncate .Bookmark.URL 100 }}</p>
          </h4>
        </a>
      </div>
      <div class="bookmark__actions media-right">
       {{ if eq .UID .Bookmark.UserID }}
        <a href="{{ BaseURL "/edit_bookmark" }}?id={{ .Bookmark.ID }}">
          <i class="fas fa-pencil-alt"></i>
        </a>
        {{ else }}
          <a href="{{ BaseURL "/bookmark" }}?id={{ .Bookmark.ID }}">
             <i class="fas fa-eye"></i>
          </a>
        {{ end }}
          <i class="fas fa-trash"></i>
          <i class="fas fa-heart"></i>
          <i class="fas fa-share-alt"></i>
      </div>
    </div>
    <div class="bookmark__tags">
        {{ if .Bookmark.Tags }}
          {{ range .Bookmark.Tags }}
            <a href="{{ if or (eq $.Page "bookmarks") (ne $.Bookmark.UserID $.UID) }}{{ BaseURL "/bookmarks" }}{{ else }}{{ BaseURL "/my_bookmarks" }}{{ end }}?tag={{ .Text }}"><span class="tag is-info">{{ .Text }}</span></a>
          {{ end }}
        {{ end }}
    </div>
    <div class="bookmark__more-info">
      <div class="bookmark__notes">
        <div class="my-bookmarks__section-header">
          <h3>
            Notes
          </h3>
        </div>
        <p class="bookmark__note">{{ .Bookmark.Notes }}</p>
      </div>
      <div class="bookmark__snapshots">
        <div class="my-bookmarks__section-header">
          <h3>
            Snapshots
          </h3>
        </div>
         {{ block "snapshots" .Bookmark.Snapshots }}{{ end }}
      </div>
    </div>
</div>
{{ end}}

{{ define "snapshots" }}
    {{ range $i,$s := . }}
    <div class="snapshot__link">
      <div>
        <a href="{{ BaseURL "/snapshot" }}?sid={{ $s.Key }}&bid={{ $s.BookmarkID }}">
          <span class="snapshot__date">{{ $s.CreatedAt | ToDate }}</span> 
          <span class="snapshot__title">
          {{if $s.Title}}{{ $s.Title }}{{else}}snapshot #{{ $i }}{{end}}
          </span>
        </a>
      </div>
      <div class="bookmark__actions">
        <i class="fas fa-pencil-alt"></i>
        <i class="fas fa-trash"></i>
      </div>
    </div>
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
