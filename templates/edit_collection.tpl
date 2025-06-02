{{ define "content" }}
<div class="content">
    <h2 class="title">{{ if .Collection }}Edit Collection{{ else }}Create Collection{{ end }}</h2>

    <form method="post" action="">
        {{ if .Collection }}
            <input type="hidden" name="cid" value="{{ .Collection.ID }}" />
        {{ end }}
        <div class="field">
            <label class="label">Name</label>
            <div class="control">
                <input class="input" type="text" name="name" value="{{ if .Collection }}{{ .Collection.Name }}{{ end }}" />
            </div>
        </div>
        {{ if .Parents }}
        <div class="field">
            <label class="label">Parent</label>
            <div class="control">
                <div class="select">
                    <select name="parent_cid">
                        <option value="0">---</option>
                        {{ $c := .Collection }}
                        {{ range .Parents }}
                            {{ if or (eq $c nil) (ne $c.ID .ID) }}
                            <option value="{{ .ID }}" {{ if and (ne $c nil) (eq $c.ParentID .ID) }}selected="selected"{{ end }}>{{ .Name }}</option>
                            {{ end }}
                        {{ end }}
                    </select>
                </div>
            </div>
        </div>
        {{ end }}
        <div class="field">
            <div class="control">
                <input class="button is-primary" type="submit" value="Save" />
            </div>
        </div>
    </form>
</div>
{{ end }}
