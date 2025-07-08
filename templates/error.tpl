{{ define "content" }}
<div class="content">
    {{ block "error" KVData "Error" .Title "Tr" .Tr "Message" .Message }}{{ end }}
</div>
{{ end }}
