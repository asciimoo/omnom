{{ define "full-content" }}
<section id="home" class="landing-page">
    <div class="landing-page__hero">
        <div class="landing-page__headline">
            <div class="columns is-desktop is-multiline is-centered is-vcentered">
                <div class="landing-page__logo column"><img src="{{ BaseURL "/static/images/omnom_logo_1024_white.png" }}" /></div>
                <div class="landing-page__title column">
                    <h1>Bookmarking and snapshotting websites made easy</h1>
                    <form method="get" action="{{ BaseURL "/signup" }}">
                        <button class="button">Sign me up!</button>
                    </form>
                </div>
            </div>
        </div>
        <div class="landing-page__menu">
        <a href="#whatis"><span class="landing-page__menu-item">What's OMNOM?</span></a>
        <a href="#extensions"><span class="landing-page__menu-item">Extensions</span></a>
        </div>
    </div>
    <div class="landing-page__content">
            <section id="whatis" class="landing-page__section">
                <div class="landing-page__article">
                    <div class="landing-page__pic">
                        <img src="{{ BaseURL "/static/placeholder-image.png" }}" />
                    </div>
                    <div class="landing-page__text">
                        <h2>What's OMNOM</h2>
                        <p>
                            Omnom is a webpage bookmarking and snapshotting service.<br />
                            Create self-contained snapshots for every bookmark you save and access or share previously visited pages without worrying about modifications or availibilty.<br />
                            Omnom consists of two parts;
                            <ul>
                                <li>A multi-user web application that accepts bookmarks & snapshots</li>
                                <li>A browser extension responsible for bookmark and snapshot creation</li>
                            </ul>
                        </p>
                    </div>
                </div>
                <a href="#extensions"><div class="next-section">
                    <i class="fas fa-angle-down"></i>
                </div></a>
            </section>
            <section id="extensions" class="landing-page__section">
                <div class="landing-page__article">
                    <div class="landing-page__pic">
                        <img src="{{ BaseURL "/static/placeholder-image.png" }}" />
                    </div>
                    <div class="landing-page__text">
                        <h2>Extensions</h2>
                        <p>
                            Browser extensions are required to create bookmarks & snapshots.<br />
                            Install the extension to your browser and enjoy Omnoming.
                        </p>
                        <h3>Get plugin</h3>
                        <div class="landing-page__extensions">
                            <form method="get" action="https://addons.mozilla.org/en-US/firefox/addon/omnom/" target="_blank">
                                <button class="extension-button">
                                    <i class="fab fa-firefox"></i>
                                    <p>Firefox</p>
                                </button>
                            </form>
                            <br />
                            <form method="get" action="https://chrome.google.com/webstore/detail/omnom/nhpakcgbfdhghjnilnbgofmaeecoojei" target="_blank">
                                <button class="extension-button">
                                    <i class="fab fa-chrome"></i>
                                    <p>Chrome</p>
                                </button>
                            </form>
                        </div>
                    </div>
                </div>
            </section>
    </div>
</section>
{{ end }}
