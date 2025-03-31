{{ define "content" }}
{{ block "warning" "This is an experimental feature, always check the output"}}{{ end }}
<div class="content">
    <h3 class="title">Create new bookmark</h3>
    <form action="" method="post">
    <div class="columns">
        <div class="column">
            <div class="field">
                <label class="label">URL</label>
                <div class="control">
                    <input class="input" type="text" placeholder="URL" name="url">
                </div>
            </div>
            <div class="field">
                <label class="label">Title</label>
                <div class="control">
                    <input class="input" type="text" placeholder="Title" name="title">
                </div>
            </div>
            <div class="field">
                <label class="label">Tags</label>
                <div class="control">
                    <input class="input" type="text" placeholder="Tags" name="tags">
                </div>
            </div>
            <div class="field">
                <label class="label">Notes</label>
                <div class="control">
                    <textarea class="textarea" name="notes"></textarea>
                </div>
            </div>
            <div class="field">
                <label class="checkbox">
                    <input type="checkbox" name="public">
                    Public
                </label>
            </div>
            <input type="hidden" name="_csrf" value="{{ .CSRF }}" />
            {{ block "submit" "Save" }}{{ end }}
        </div>
    </div>
    </form>
</div>
{{ end }}
