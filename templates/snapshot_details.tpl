{{ define "content" }}
<div class="content">
    <h4 class="title">
        Snapshot Details: <a href="{{ .URL }}">{{ .URL }}</a>
    </h4>
    <hr />
    <div class="columns has-text-centered is-vcentered">
        <div class="column">
            <span class="tag is-info is-large">{{ .Snapshot.Size | FormatSize }}</span>
            <br />Total Size
        </div>
        <div class="column">
            <span class="tag is-info is-large">{{ .ResourceCount }}</span>
            <br />Total resources
        </div>
    </div>
    <h4 class="title">
        Resource List
    </h4>
    <div class="grid is-col-min-18 resources">
    {{ range $k, $t := .Resources }}
        <div class="cell">
            <nav class="panel">
                <p class="panel-heading">{{ $k | Capitalize }}</p>
                {{ range $sk, $l := $t }}
                <h5 class="m-4">{{ $sk }} ({{ len $l }})</h5>
                {{ range $l }}<a href="{{ .Key | ResourceURL }}" class="panel-block">{{ .OriginalFilename }} <span class="tag">{{ .Size | FormatSize }}</span></a>{{ end }}
                {{ end }}
            </nav>
        </div>
    {{ end }}
    </div>
</div>
{{ end }}
