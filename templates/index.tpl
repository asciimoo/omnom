{{ define "full-content" }}
<section id="home" class="landing-page">
    <div class="landing-page__hero">
        <div class="landing-page__headline">
            <div class="landing-page__title">
                <h1>Bookmarking and snapshotting websites made easy</h1>
                <button class="button">Sign me up!</button>
            </div>
        </div>
        <div class="landing-page__menu">
        <span class="landing-page__menu-item">What's OMNOM?</span>
        <span class="landing-page__menu-item">Extensions</span>
        </div>
    </div>
    <div class="landing-page__content">
            <section class="landing-page__section">
                <div class="landing-page__article">
                    <div class="landing-page__pic">
                        <img src="{{ BaseURL "/static/placeholder-image.png" }}" />
                    </div>
                    <div class="landing-page__text">
                        <h2>What's OMNOM</h2>
                        <p>Loremimpsum LoremimpsumLoremimpsumLoremimpsum Loremimpsum LoremimpsumLoremimpsumLoremimpsum Loremimpsum LoremimpsumLoremimpsumLoremimpsum Loremimpsum LoremimpsumLoremimpsum</p>
                    </div>
                </div>
                <div class="next-section">
                    <i class="fas fa-angle-down"></i>
                </div> 
            </section>
            <section class="landing-page__section">
                <div class="landing-page__article">
                    <div class="landing-page__pic">
                        <img src="{{ BaseURL "/static/placeholder-image.png" }}" />
                    </div>
                    <div class="landing-page__text">
                        <h2>Extensions</h2>
                        <p>Loremimpsum LoremimpsumLoremimpsumLoremimpsum Loremimpsum LoremimpsumLoremimpsumLoremimpsum Loremimpsum LoremimpsumLoremimpsumLoremimpsum Loremimpsum LoremimpsumLoremimpsum</p>
                        <h3>Get plugin</h3>
                        <div class="landing-page__extensions">
                            <button class="extension-button">
                                <i class="fab fa-firefox"></i>
                                <p>Firefox</p>
                            </button>
                            <button class="extension-button">
                                <i class="fab fa-chrome"></i>
                                <p>Chrome</p>
                            </button>
                        </div>
                    </div>
                </div>
                <div class="next-section">
                    <i class="fas fa-angle-down"></i>
                </div> 
            </section>   
    </div>
</section>
{{ end }}
