{{ define "full-content" }}
<div class="hero is-medium">
    <div class="hero-body shadow-bottom">
        <div class="columns">
            <div class="column has-text-centered">
                <h2 class="title has-text-white">
                    {{ .Tr.Msg "landing page slogan" | ToHTML }}
                </h2>
                <p class="big has-text-info">
                    {{ .Tr.Msg "landing page subslogan 1" }}<br/>{{ .Tr.Msg "landing page subslogan 2" }}
                </p>
            </div>
        </div>
        <div class="columns has-text-centered is-centered m-5">
            <div class="column">
                <div class="field is-grouped has-addons has-addons-centered">
                    {{ if not .DisableSignup }}
                    <div class="control">
                        <form method="get" action="{{ URLFor "Signup" }}">
                            <button class="button is-link is-large">{{ .Tr.Msg "sign up" }}</button>
                        </form>
                    </div>
                    {{ end }}
                    <div class="control">
                        <a href="{{ URLFor "Public bookmarks" }}" class="button is-link is-large">{{ .Tr.Msg "public bookmarks" }}</a>
                    </div>
                </div>
            </div>
        </div>
        <div class="columns landing-features is-centered">
            <div class="column is-one-fifth-fullhd">
                <div class="box has-text-centered is-maxheight">
                    <header class="is-size-3">
                        <span class="icon is-large has-text-info"><i class="fa-solid fa-camera-retro"></i></span><br/>
                        {{ .Tr.Msg "landing feature 1 title" }}
                    </header>
                    <div class="content has-text-info">
                        {{ .Tr.Msg "landing feature 1 desc" }}
                    </div>
                </div>
            </div>
            <div class="column is-one-fifth-fullhd">
                <div class="box has-text-centered is-maxheight">
                    <header class="is-size-3">
                        <span class="icon is-large has-text-info"><i class="fa-regular fa-shield"></i></span><br/>
                        {{ .Tr.Msg "landing feature 2 title" }}
                    </header>
                    <div class="content has-text-info">
                        {{ .Tr.Msg "landing feature 2 desc" }}
                    </div>
                </div>
            </div>
            <div class="column is-one-fifth-fullhd">
                <div class="box has-text-centered is-maxheight">
                    <header class="is-size-3">
                        <span class="icon is-large has-text-info"><i class="fa-solid fa-folder"></i></span><br/>
                        {{ .Tr.Msg "landing feature 3 title" }}
                    </header>
                    <div class="content has-text-info">
                        {{ .Tr.Msg "landing feature 3 desc" }}
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>
<div class="container mt-6 has-text-centered">
    <div class="columns is-centered m-5">
        <div class="column">
            <h3 class="title is-2">{{ .Tr.Msg "landing subslogan 3 title" }}</h3>
            <p class="title is-5">{{ .Tr.Msg "landing subslogan 3 desc" }}</p>
        </div>
    </div>
    <div class="columns is-centered my-6 mx-1">
        <div class="column">
            <div class="box has-text-centered is-maxheight">
                <header class="is-size-3">
                    <span class="icon is-large has-text-primary"><i class="fa-solid fa-share-nodes"></i></span><br/>
                    {{ .Tr.Msg "landing feature 4 title" }}
                </header>
                <div class="content">
                    {{ .Tr.Msg "landing feature 4 desc" }}
                </div>
            </div>
        </div>
        <div class="column">
            <div class="box has-text-centered is-maxheight">
                <header class="is-size-3">
                    <span class="icon is-large has-text-primary"><i class="fa-regular fa-search"></i></span><br/>
                    {{ .Tr.Msg "landing feature 5 title" }}
                </header>
                <div class="content">
                    {{ .Tr.Msg "landing feature 5 desc" }}
                </div>
            </div>
        </div>
        <div class="column">
            <div class="box has-text-centered is-maxheight">
                <header class="is-size-3">
                    <span class="icon is-large has-text-primary"><i class="fa-solid fa-code-compare"></i></span><br/>
                    {{ .Tr.Msg "landing feature 6 title" }}
                </header>
                <div class="content">
                    {{ .Tr.Msg "landing feature 6 desc" }}
                </div>
            </div>
        </div>
    </div>
    <div class="box mb-6">
        <h3 class="title is-3 mt-6">{{ .Tr.Msg "download extension" }}</h3>
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
</div>
{{ end }}
