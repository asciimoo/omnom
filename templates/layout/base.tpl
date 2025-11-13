<!DOCTYPE html>
<html lang="en" data-theme="{{ .Theme }}">
<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Omnom</title>
    <link rel="stylesheet" href="{{ BaseURL "/static/css/bulma.css" }}" />
    <link rel="stylesheet" href="{{ BaseURL "/static/css/style.css" }}" />
    <link rel="icon" type="image/png" href="{{ BaseURL "/static/icons/omnom.png" }}" sizes="128x128">

    {{ block "head" . }} {{ end }}
</head>
<body id="omnom-webapp">
<input type="hidden" id="has_user" value="{{ if .User }}1{{ else }}0{{ end }}" />
<input type="hidden" id="base_url" value="{{ BaseURL "/" }}" />
<div id="modal" class="modal"></div>
<div class="webapp__content {{ block "content-class" . }}{{ end }}">
<nav class="navbar {{ block "content-class" . }}{{ end }}" role="navigation" aria-label="main navigation">
  <div class="navbar__container{{ if ne .Page "index" }} shadow-bottom{{ end }}">
    <div class="navbar-brand is-size-4">
      <a class="navbar__logo" href="{{ URLFor "Index" }}"><span>om</span><span class="text--primary">nom</span> </a>
      <label for="nav-toggle-state" role="button" class="navbar-burger burger has-text-black" aria-label="menu" aria-expanded="false">
        <span aria-hidden="true"></span>
        <span aria-hidden="true"></span>
        <span aria-hidden="true"></span>
        <span aria-hidden="true"></span>
      </label>
    </div>
    <input type="checkbox" id="nav-toggle-state" />

    <div id="navbar-menu" class="navbar-menu is-size-5">
        <div class="navbar-start">
        {{ if not (eq .Page "index") }}
            <a href="{{ URLFor "Index" }}" class="navbar-item{{ if or (eq .Page "index") (eq .Page "dashboard") }} is-active{{ end }}">{{ .Tr.Msg "home" }}</a>
            {{ if .User }}
            <a href="{{ URLFor "My bookmarks" }}" class="navbar-item{{ if eq .Page "my-bookmarks" }} is-active{{ end }}">{{ .Tr.Msg "my bookmarks" }}</a>
            <a href="{{ URLFor "Feeds" }}" class="navbar-item{{ if eq .Page "feeds" }} is-active{{ end }}">{{ .Tr.Msg "feeds" }}</a>
            {{ end }}
            <a href="{{ URLFor "Public bookmarks" }}" class="navbar-item{{ if eq .Page "bookmarks" }} is-active{{ end }}">{{ .Tr.Msg "public bookmarks" }}</a>
            {{ if .User }}
            <a href="{{ URLFor "Create bookmark form" }}" class="navbar-item{{ if eq .Page "create-bookmark" }} is-active{{ end }}">{{ .Tr.Msg "create bookmark" }}</a>
            {{ end }}
        {{ end }}
      </div>
      <div class="navbar-end">
        <form action="{{ URLFor "search" }}">
            <div class="field is-horizontal navbar-item">
                <div class="field-body">
                    <div class="field">
                        <div class="control has-icons-left">
                            <input class="input is-outlined is-info" type="search" placeholder="{{ .Tr.Msg "search" }}" name="q" value="{{ or .SearchParams.Q .Query }}" id="search-input">
                            <span class="icon is-small is-left has-text-info">
                            <i class="fas fa-search"></i>
                            </span>
                        </div>
                    </div>
                </div>
            </div>
        </form>
        {{ if .User }}
            <a href="{{ URLFor "Profile" }}" class="navbar-item"><i class="fas fa-user"></i> &nbsp; {{ .User.Username }}</a>
              {{ if .AllowManualLogin }}
                <div class="navbar-item"><a href="{{ URLFor "Logout" }}" class="button is-outlined is-info">{{ .Tr.Msg "logout" }}</a></div>
              {{ end }}
        {{ else if .AllowManualLogin }}
            <div class="navbar-item"><a href="{{ URLFor "Login" }}" class="button is-outlined is-info">{{ .Tr.Msg "login" }}</a></div>
            {{ if not .DisableSignup }}<div class="navbar-item"><a href="{{ URLFor "Signup" }}" class="button is-outlined is-info">{{ .Tr.Msg "sign up" }}</a></div>{{ end }}
        {{ end }}
      </div>
    </div>
  </div>
</nav>

    {{ if .Error }}
    <div class="section">{{ block "error" . }}{{ end }}</div>
    {{ end }}
    {{ if .Warning }}
    <div class="section">{{ block "warning" . }}{{ end }}</div>
    {{ end }}
    {{ if .Info }}
    <div class="section">{{ block "info" . }}{{ end }}</div>
    {{ end }}
{{block "full-content" . }}
<div class="section webapp__main-container">
    <div class="bd-main-container container">
        {{ block "content" . }}{{ end }}
    </div>
</div>
{{ end }}
{{ if (not .hideFooter) }}
<footer class="footer">
    <div class="container px-6">
        <div class="columns is-centered">
            <div class="column">
                <img src="{{ BaseURL "/static/icons/omnom-logo-v1.svg" }}" class="icon is-large" />
                <p>Omnom © 2025</p>
            </div>
            <div class="column is-narrow px-6">
                <p><b>{{ .Tr.Msg "product" }}</b></p>
                {{ if eq .Theme "dark" }}<a href="?theme=light">{{ .Tr.Msg "light theme" }}</a>{{ else }}<a href="?theme=dark">{{ .Tr.Msg "dark theme" }}</a>{{ end }}
                <br /><a href="https://addons.mozilla.org/en-US/firefox/addon/omnom/">{{ .Tr.Msg "firefox ext" }}</a>
                <br /><a href="https://chrome.google.com/webstore/detail/omnom/nhpakcgbfdhghjnilnbgofmaeecoojei">{{ .Tr.Msg "chrome ext" }}</a>
                <br /><a href="{{ AddURLParam .URL "format=json" }}">{{ .Tr.Msg "json view" }}</a>
            </div>
            <div class="column is-narrow pl-6">
                <p><b>{{ .Tr.Msg "support" }}</b></p>
                <a href="{{ URLFor "API" }}">API</a>
                <br /><a href="https://github.com/asciimoo/omnom">GitHub</a>
                <br /><a href="https://github.com/asciimoo/omnom/wiki">Wiki</a>
            </div>
        </div>
    </div>
</footer>
{{ end }}
</div>
<template id="tpl-modal">
    <div class="modal-background"></div>
    <div class="modal-content card p-5">
        <h2 class="title is-4">{{/* .Tr.Msg "XY" */}}</h2>
        <button class="button tpl-yes is-large">{{/* .Tr.Msg "Yes" */}}</button>
        <button class="button tpl-no is-large" aria-label="close">{{/* .Tr.Msg "No" */}}</button>
    </div>
</template>
<script src="{{ BaseURL "/static/js/site.js" }}"></script>
</body>
</html>

{{ define "error" }}
<article class="message is-danger container is-size-5">
  <div class="message-header">
    <p>{{ .Tr.Msg "error" }}</p>
  </div>
  <div class="message-body">
      <b>{{ .Error | ToHTML }}</b>
      {{ if .Message }}
      <p>{{ .Message }}</p>
      {{ end }}
  </div>
</article>
{{ end }}

{{ define "warning" }}
<article class="message is-warning container is-size-5">
  <div class="message-header">
    <p>{{ .Tr.Msg "warning" }}</p>
  </div>
  <div class="message-body">{{ .Warning | ToHTML }}</div>
</article>
{{ end }}

{{ define "info" }}
<article class="message is-info container is-size-5">
  <div class="message-body">{{ .Info | ToHTML }}</div>
</article>
{{ end }}


{{ define "note" }}
<article class="message is-info container is-size-5">
  <div class="message-header">
    <p>{{ .Tr.Msg "note" }}</p>
  </div>
  <div class="message-body">{{ .Note | ToHTML }}</div>
</article>
{{ end }}

{{ define "collections" }}
{{ if .Collections }}
<ul>
    {{ $cc := .CurrentCollection }}
    {{ $tr := .Tr }}
    {{ range .Collections }}
    <li>
        <a href="{{ URLFor "my bookmarks" }}?collection={{ .Name }}" {{ if eq $cc .Name }}class="has-background-info-light"{{ end }}><span class="icon"><i class="fas fa-folder"></i></span>{{ .Name }}</a>
        <div class="is-pulled-right">
            <a href="{{ URLFor "edit collection form" }}?cid={{ .ID }}"><span class="icon"><i class="fas fa-pencil-alt"></i></span></a>
        </div>
    </li>
    {{ block "collections" KVData "Collections" .Children "CurrentCollection" $cc "Tr" $tr }}{{ end }}
    {{ end }}
</ul>
{{ end }}
{{ end }}

{{ define "bookmark" }}
<div class="media bookmark__container">
    <div class="bookmark__header">
      <div class="bookmark__title">
        <div class="bookmark__favicon">
            <span class="icon">
            {{ if .Bookmark.Favicon }}
              <img src="{{ .Bookmark.Favicon | ToURL }}" alt="{{ .Tr.Msgf "favicon of" "Title" .Bookmark.Title }}" />
            {{ end }}
            </span>
        </div>
          <h4 class="title">
              <a href="{{ .Bookmark.URL }}" target="_blank">
                {{ .Bookmark.Title }}
              </a>
              <p class="is-size-7 has-text-grey has-text-weight-normal">
                  {{ Truncate .Bookmark.URL 100 }}<br />
                  <span class="has-text-black">{{ .Bookmark.CreatedAt | ToDate }}</span>
                  <a href="{{ URLFor "User" .Bookmark.User.Username }}">@{{ .Bookmark.User.Username }}</span></a>
                  {{ if and (eq .Bookmark.UserID $.UID) (ne .Bookmark.Collection nil) .Bookmark.Collection.ID }}
                  <a href="{{ URLFor "my bookmarks" }}?collection={{ .Bookmark.Collection.Name }}" title="{{ .Tr.Msg "collection" }}"><span class="icon"><i class="fas fa-folder"></i></span>{{ .Bookmark.Collection.Name }}</a>
                  {{ end }}
                  {{ if .Bookmark.Tags }}
                  <span class="bookmark__tags">
                      {{ range .Bookmark.Tags }}
                      <a href="{{ if or (eq $.Page "bookmarks") (ne $.Bookmark.UserID $.UID) }}{{ URLFor "Public bookmarks" }}{{ else }}{{ URLFor "My bookmarks" }}{{ end }}?tag={{ .Text }}"><span class="tag is-muted-primary">{{ .Text }}</span></a>
                      {{ end }}
                  </span>
                  {{ end }}
              </p>
          </h4>
      </div>
      <div class="bookmark__actions">
          {{ if .Bookmark.Snapshots }}
          {{ if eq (len .Bookmark.Snapshots) 1 }}
          <a href="{{ URLFor "snapshot" }}?bid={{ .Bookmark.ID }}&sid={{ (index .Bookmark.Snapshots 0).Key }}">{{ .Tr.Msg "snapshot" }}</a>
          {{ else }}
          <details>
              <summary>
                  Snapshots <span class="bookmark__snapshot-count">({{len .Bookmark.Snapshots}})</span>
              </summary>
              <div>
                  {{ block "snapshots" KVData "Snapshots" .Bookmark.Snapshots "IsOwn" (eq .Bookmark.UserID .UID) }}{{ end }}
              </div>
          </details>
          {{ end }}
          {{ end }}
        {{ if .Bookmark.Notes }}
        <details>
          <summary>
            Notes
          </summary>
          <div>
              <p class="has-text-black">{{ .Bookmark.Notes }}</p>
          </div>
        </details>
        {{ end }}
          <span class="tag">{{ if .Bookmark.Public }}{{ .Tr.Msg "public" }}{{ else }}{{ .Tr.Msg "private" }}{{ end }}</span>
          <a href="{{ URLFor "Bookmark" }}?id={{ .Bookmark.ID }}" title="{{ .Tr.Msg "view" }}"><i class="fas fa-eye"></i></a>
          {{ if eq .UID .Bookmark.UserID }}
          <a href="{{ URLFor "Edit bookmark" }}?id={{ .Bookmark.ID }}" title="{{ .Tr.Msg "edit" }}"><i class="fas fa-pencil-alt"></i></a>
          {{ end }}
          <!--<i class="fas fa-heart"></i>
          <i class="fas fa-share-alt"></i>-->
      </div>
    </div>
</div>
{{ end}}

{{ define "snapshots" }}
    {{ range $i,$s := .Snapshots }}
    <div class="snapshot__link">
      <div>
        <a href="{{ URLFor "Snapshot" }}?sid={{ $s.Key }}&bid={{ $s.BookmarkID }}">
          <span class="snapshot__date">{{ $s.CreatedAt | ToDate }}</span>
          <span class="snapshot__title">
          {{if $s.Title}}{{ $s.Title }}{{else}}snapshot #{{ $i }}{{end}}
          </span>
        </a>
      </div>
      <div class="bookmark__actions">
          {{ if $.IsOwn }}
          <i class="fas fa-pencil-alt"></i>
          <form method="post" action="{{ URLFor "Delete snapshot" }}">
              <input type="hidden" name="bid" value="{{ $s.BookmarkID }}" />
              <input type="hidden" name="sid" value="{{ $s.ID }}" />
              <button class="snapshot__delete" type="submit">
                  <i class="fas fa-trash"></i>
              </button>
          </form>
          {{ end }}
      </div>
    </div>
    {{ end }}
{{ end }}

{{ define "paging" }}
<div class="columns is-centered">
    <div class="column is-narrow">
        {{ if and .Pageno (gt .Pageno 1) }}
        <a href="{{ AddURLParam .URL (printf "pageno=%d" (dec .Pageno)) }}" class="button is-primary is-medium"><span class="icon"><i class="fas fa-angle-left"></i></span><span>{{ .Tr.Msg "previous page" }}</span></a>
        {{ end }}
        {{ if .HasNextPage }}
        <a href="{{ AddURLParam .URL (printf "pageno=%d" (inc .Pageno)) }}" class="button is-primary is-medium"><span>{{ .Tr.Msg "next page" }}</span><span class="icon"><i class="fas fa-angle-right"></i></span></a>
        {{ end }}
    </div>
</div>
{{ end }}

{{ define "textFilter" }}
<div class="field is-horizontal">
    <div class="field-body">
        <div class="field">
            <div class="control has-icons-left">
                <input class="input" type="search" placeholder="{{ .Tr.Msg "search" }}" name="query" value="{{ or .SearchParams.Q .Query }}">
                 <span class="icon is-small is-left">
                <i class="fas fa-search"></i>
                </span>
            </div>
        </div>
    </div>
</div>
{{ end }}

{{ define "searchParameters"}}
<div class="checkboxes">
    <label class="label" for="search_in_snapshot">
        <input class="switch is-rounded" value="true" type="checkbox" id="search_in_snapshot"  name="search_in_snapshot"{{ if .SearchParams.SearchInSnapshot }} checked="checked"{{ end }}>
        {{ .Tr.Msg "snapshot content search" }}
    </label>
    <label class="label">
        <input class="switch is-rounded" value="true" type="checkbox" id="search_in_note" name="search_in_note"{{ if .SearchParams.SearchInNote }} checked="checked"{{ end }}>
        {{ .Tr.Msg "note search" }}
    </label>
    {{ if eq .Page "my-bookmarks" }}
    <label class="label">
        <input class="switch is-rounded" value="true" type="checkbox" id="public" name="public"{{ if .SearchParams.IsPublic }} checked="checked"{{ end }}>
        {{ .Tr.Msg "only public bm" }}
    </label>
    {{ end }}
</div>
{{ end }}

{{ define "domainFilter" }}
<div class="field">
    <label class="label">{{ .Tr.Msg "domain" }}</label>
    <div class="control">
        <input class="input" type="search" placeholder="{{ .Tr.Msg "domain" }}.." name="domain" value="{{ .SearchParams.Domain }}">
    </div>
</div>
{{ end }}

{{ define "ownerFilter" }}
<div class="field">
<label class="label">{{ .Tr.Msg "owner" }}</label>
    <div class="control">
        <input class="input" type="search" placeholder="{{ .Tr.Msg "username" }}.." name="owner" value="{{ .SearchParams.Owner }}">
    </div>
</div>
{{ end }}

{{ define "tagFilter" }}
<div class="field">
<label class="label">{{ .Tr.Msg "tags" }}</label>
    <div class="control">
        <input class="input" type="search" placeholder="{{ .Tr.Msg "tag" }}.." name="tag" value="{{ .SearchParams.Tag }}">
    </div>
</div>
{{ end }}

{{ define "collectionFilter" }}
{{ if .Collections }}
<div class="field">
    <label class="label">{{ .Tr.Msg "collection" }}</label>
    <div class="control">
        <div class="select">
            {{ $cid := .CurrentCollection }}
            <select name="collection">
                <option value="">---</option>
                {{ range .Collections }}
                <option value="{{ .Name }}" {{ if eq .Name $cid }}selected="selected"{{ end }}>{{ .Name }}</option>
                {{ range .Children }}
                <option value="{{ .Name }}" {{ if eq .Name $cid }}selected="selected"{{ end }}>{{ .Name }}</option>
                {{ end }}
                {{ end }}
            </select>
        </div>
    </div>
</div>
{{ end }}
{{ end }}

{{ define "dateFilter" }}
<div class="field is-grouped is-grouped-multiline">
    <div class="control">
    <div class="field">
        <label class="label">{{ .Tr.Msg "date from" }}</label>
            <p class="control is-expanded">
                <input class="input" type="date" placeholder="YYYY.MM.DD" name="from" value="{{ .SearchParams.FromDate }}">
            </p>
        </div>
    </div>
    <div class="control">
        <div class="field">
        <label class="label">{{ .Tr.Msg "date to" }}</label>
            <p class="control is-expanded">
                <input class="input" type="date" placeholder="YYYY.MM.DD" name="to" value="{{ .SearchParams.ToDate }}">
            </p>
        </div>
    </div>
</div>
{{ end }}


{{ define "sortBy" }}
<span class="select">
    <select name="order_by">
        <option value="date_desc"{{ if eq .OrderBy "date_desc" }} selected="selected"{{ end }}>{{ .Tr.Msg "date desc" }}</option>
        <option value="date_asc"{{ if eq .OrderBy "date_asc" }} selected="selected"{{ end }}>{{ .Tr.Msg "date asc" }}</option>
    </select>
</span>
<input type="submit" value="{{ .Tr.Msg "sort" }}" class="button" />
{{ end }}

{{ define "submit" }}
<div class="control">
    <input type="submit" name="submit" value="{{ . }}" class="button is-primary" />
</div>
{{ end }}

{{ define "feedItem" }}
    <div class="media">
        <div class="media-left">
            <figure class="image is-48x48">
                {{ if .Item.Favicon }}
                <img src="{{ .Item.Favicon | ToURL }}" alt="favicon" />
                {{ end }}
            </figure>
        </div>
        <div class="media-content">
            {{ if .Item.Unread }}
            <div class="is-pulled-right"><form method="post" action="{{ URLFor "archive items" }}"><input type="hidden" name="fids" value="{{ .Item.UserFeedItemID }}"><input type="submit" class="button is-info" value="{{ .Tr.Msg "archive item" }}"></form></div>
            {{ end }}
            <p class="title is-5">
                <a href="{{ .Item.URL }}">{{ .Item.Title }}</a>
            </p>
            <p class="subtitle is-6">
                <span class="tag">{{ .Item.FeedName }}</span> {{ .Item.CreatedAt | ToDateTime }}
                {{ if ne .Item.OriginalAuthor "" }}
                <br /><b>Reposted from <a href="{{ .Item.OriginalAuthor }}">{{ .Item.OriginalAuthor }}</a></b>
                {{ end }}
            </p>
            {{ if .Item.Content }}
            <article class="{{ .Item.FeedType }} content">{{ .Item.Content | ToHTML }}</article>
            {{ end }}
        </div>
    </div>
{{ end }}

{{ define "feedBookmarkItem" }}
    <div class="media">
        <div class="media-left">
            <figure class="image is-48x48">
            {{ if .Bookmark.Favicon }}
              <img src="{{ .Bookmark.Favicon | ToURL }}" alt="favicon" />
            {{ end }}
            </figure>
        </div>
        <div class="media-content">
            {{ if .Bookmark.Unread }}
            <div class="is-pulled-right"><form method="post" action="{{ URLFor "archive items" }}"><input type="hidden" name="bids" value="{{ .Bookmark.ID }}"><input type="submit" class="button is-info" value="{{ .Tr.Msg "archive item" }}"></form></div>
            {{ end }}
            <p class="title is-5"><a href="{{ .Bookmark.URL }}">{{ .Bookmark.Title }}</a></p>
            <p class="subtitle is-6"><span class="tag is-muted-primary">{{ .Tr.Msg "bookmark" }}</span> {{ .Bookmark.CreatedAt | ToDateTime }}</p>
        </div>
    </div>
    {{ if .Bookmark.Notes }}
    <p>{{ .Bookmark.Notes }}</p>
    {{ end }}
{{ end }}

{{ define "feedSidebar" }}
<div class="column is-2-fullhd is-one-quarter-desktop is-one-third-tablet">
    <div class="content">
        {{ if not .Feeds }}
        <h3 class="title">{{ .Tr.Msg "no feeds found" }}</h3>
        {{ end }}
        <form action="{{ URLFor "feed search" }}" method="get">
            {{ if .FeedID }}<input type="hidden" name="feed_id" value="{{ .FeedID }}" />{{ end }}
            {{ block "textFilter" . }}{{ end }}
            <div class="checkboxes">
                <label class="label" for="include_read_items">
                    <input class="switch is-rounded" value="true" type="checkbox" id="include_read_items"  name="include_read_items"{{ if .IncludeRead }} checked="checked"{{ end }}>
                    {{ .Tr.Msg "include read items" }}
                </label>
            </div>
            {{ block "submit" (.Tr.Msg "search") }}{{ end }}
        </form>
        <details class="my-4 is-size-4">
            <summary>{{ .Tr.Msg "add feed" }}</summary>
            <form action="{{ URLFor "add feed" }}" method="post">
                <div class="field">
                    <label class="label">{{ .Tr.Msg "name" }}</label>
                    <div class="control">
                        <input class="input" type="text" placeholder="{{ .Tr.Msg "name" }}.." name="name" />
                    </div>
                </div>
                <div class="field">
                    <label class="label">{{ .Tr.Msg "url" }}</label>
                    <div class="control">
                        <input class="input" type="text" placeholder="{{ .Tr.Msg "url" }}.." name="url" />
                    </div>
                </div>
                {{ block "submit" (.Tr.Msg "submit") }}{{ end }}
            </form>
        </details>
        {{ $Tr := .Tr }}
        {{ $IncludeRead := .IncludeRead }}
        {{ range .Feeds }}
        <h4>
            <div class="is-pulled-right">
                <a href="{{ URLFor "edit feed" }}?id={{ .ID }}" aria-label="{{ $Tr.Msg "edit feed" }}"><span class="icon"><i class="fas fa-pencil"></i></span></a>
            </div>
            <a href="{{ URLFor "feed search" }}?feed_id={{ .ID }}{{ if $IncludeRead }}&include_read_items=1{{ end }}">{{ .Name }}</a>{{ if .Count }} <span class="tag is-medium">{{ .Count }}</span>{{ end }}
        </h4>
        {{ end }}
    </div>
</div>
{{ end }}
