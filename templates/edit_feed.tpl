{{ define "content" }}
<div class="content">
    <h2 class="title">Edit Feed</h2>

    <form method="post" action="">
        <input type="hidden" name="id" value="{{ .Feed.ID }}" />
        <div class="field">
            <label class="label">Name</label>
            <div class="control">
                <input class="input" type="text" name="name" value="{{ .Feed.Name }}" />
            </div>
        </div>
        <div class="field">
            <div class="control">
                <input class="button is-primary" type="submit" value="{{ .Tr.Msg "save" }}" />
            </div>
        </div>
    </form>
    <div class="field is-grouped is-grouped-right">
        <form method="post" action="{{ URLFor "Delete feed" }}">
            <input class="button is-danger" type="submit" value="{{ .Tr.Msg "delete feed" }}" />
            <input type="hidden" name="id" value="{{ .Feed.ID }}" />
        </form>
    </div>
</div>
{{ end }}
