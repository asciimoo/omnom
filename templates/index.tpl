{{ define "full-content" }}
<div class="hero is-medium">
    <div class="hero-body">
        <div class="columns">
            <div class="column is-2 has-text-right-tablet has-text-centered-mobile">
                <img src="{{ BaseURL "/static/images/omnom_logo_1024_white.png" }}" class="logo" />
            </div>
            <div class="column is-10">
                <h2 class="title is-2">
                    {{ .Tr.Msg "landing page slogan" }}
                </h2>
                <p class="big">
                    {{ .Tr.Msg "landing page subslogan 1" }}<br/>{{ .Tr.Msg "landing page subslogan 2" }}
                </p>
            </div>
        </div>
        <div class="columns">
            <div class="column is-offset-2 is-10">
            {{ if not .DisableSignup }}
                <form method="get" action="{{ URLFor "Signup" }}">
                    <button class="button is-link is-large">{{ .Tr.Msg "sign up" }}</button>
                </form>
            {{ end }}
            </div>
        </div>
    </div>
</div>
<div class="container mt-6 has-text-centered">
            <h3 class="title is-3">{{ .Tr.Msg "download extension" }}</h3>
            <p>
                {{ .Tr.Msg "landing page ext desc" }}
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
{{ end }}
