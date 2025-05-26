{{ define "content" }}
{{ block "warning" KVData "Warning" (.Tr.Msg "experimental feature") "Tr" .Tr }}{{ end }}
<div class="content">
    <h3 class="title">{{ .Tr.Msg "create bookmark" }}</h3>
    <form action="" method="post">
    <div class="columns">
        <div class="column">
            <div class="field">
                <label class="label">{{ .Tr.Msg "url" }}</label>
                <div class="control">
                    <input class="input" type="text" placeholder="URL" name="url">
                </div>
            </div>
            <div class="field">
                <label class="label">{{ .Tr.Msg "title" }}</label>
                <div class="control">
                    <input class="input" type="text" placeholder="Title" name="title">
                </div>
            </div>
            <div class="field">
                <label class="label">{{ .Tr.Msg "tags" }}</label>
                <div class="control">
                    <input class="input" type="text" placeholder="Tags" name="tags">
                </div>
            </div>
            <div class="field">
                <label class="label">{{ .Tr.Msg "notes" }}</label>
                <div class="control">
                    <textarea class="textarea" name="notes"></textarea>
                </div>
            </div>
            <div class="field">
                <label class="checkbox">
                    <input type="checkbox" name="public">
                    {{ .Tr.Msg "public" }}
                </label>
            </div>
            <input type="hidden" name="_csrf" value="{{ .CSRF }}" />
            {{ block "submit" (.Tr.Msg "save") }}{{ end }}
        </div>
    </div>
    </form>
</div>
{{ end }}
