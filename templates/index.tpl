{{ define "full-content" }}
<div class="hero is-medium">
    <div class="hero-body">
        <div class="columns">
            <div class="column is-2 has-text-right">
                <img src="{{ BaseURL "/static/images/omnom_logo_1024_white.png" }}" class="logo" />
            </div>
            <div class="column is-10">
                <h2 class="title is-2">
                    Bookmarking and snapshotting websites made easy
                </h2>
                <p class="big">
                    Create self-contained snapshots for every bookmark you save.<br/> Access &amp; share previously visited pages without worrying about modifications or availibilty.
                </p>
            </div>
        </div>
        <div class="columns">
            <div class="column is-offset-2 is-10">
                <form method="get" action="{{ URLFor "Signup" }}">
                    <button class="button is-link is-large">Sign me up</button>
                </form>
            </div>
        </div>
    </div>
</div>
<div class="container mt-6">
    <div class="columns has-text-centered">
        <div class="column is-12">
            <h3 class="title is-3">Download extension</h3>
            <p>
                Browser extensions are required to create bookmarks & snapshots. Install the extension to your browser and enjoy Omnoming.
            </p>
            <div class="landing-page__extensions">
                <div class="buttons is-centered">
                    <form method="get" action="https://addons.mozilla.org/en-US/firefox/addon/omnom/" target="_blank">
                        <button class="extension-button">
                            <i class="fab fa-firefox"></i>
                            <p>Firefox</p>
                        </button>
                    </form>
                    <form method="get" action="https://chrome.google.com/webstore/detail/omnom/nhpakcgbfdhghjnilnbgofmaeecoojei" target="_blank">
                        <button class="extension-button">
                            <i class="fab fa-chrome"></i>
                            <p>Chrome</p>
                        </button>
                    </form>
                </div>
            </div>
        </div>
    </div>
</div>
{{ end }}
