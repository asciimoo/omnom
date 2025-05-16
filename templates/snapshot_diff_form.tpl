{{ define "content" }}
<div class="content">
    <h2 class="title">Snapshot diff</h2>
    <form method="get" action="{{ URLFor "snapshot diff" }}">
    <div class="field">
        <label class="label">First snapshot key</label>
        <div class="control">
            <input class="input" type="text" name="s1" placeholder="First snapshot">
        </div>
    </div>
    <div class="field">
        <label class="label">Second snapshot key</label>
        <div class="control">
            <input class="input" type="text" name="s2" placeholder="Second snapshot">
        </div>
    </div>
    <div class="field">
        <div class="control">
            <button class="button is-link">Submit</button>
        </div>
    </div>
    </form>
</div>
{{ end }}
