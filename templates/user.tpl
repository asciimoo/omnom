{{ define "content" }}
<div class="media-content block m-5">
    <h2 class="title is-2">{{ .User.Username }}</h2>
    <h3 class="subtitle has-text-grey"><i>{{ .FediName }}</i></h3>
    <nav class="level">
        <div class="level-item has-text-centered">
            <div>
                <p class="title is-size-1 has-text-primary">{{ .BookmarkCount }}</p>
                <p class="heading">Bookmarks</p>
            </div>
        </div>
        <div class="level-item has-text-centered">
            <div>
                <p class="title is-size-1 has-text-primary">{{ .FollowerCount }}</p>
                <p class="heading">Followers</p>
            </div>
        </div>
    </nav>
    <div class="card">
        <div class="card-footer">
            <a href="{{ URLFor "Public bookmarks" }}?owner={{ .User.Username }}" class="card-footer-item">View bookmarks</a>
        </div>
    </div>
</div>
{{ end }}
