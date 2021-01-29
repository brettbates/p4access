action: RESPOND
message:  "
Possible ways to get access are listed below. This is a beta, please report issues to support.
{{ if .Context }}
Info:  {{ .Context }} 
{{ end }}
Groups:
{{ range $group := .Groups }}
    ----
    Group {{ $group.Group }} grants {{ $group.Access }} access to the path: 

        {{ $group.Path }}

    You can get access by contacting one of the owners listed: 
    {{ range $group.Owners }} 
        {{ .FullName }}: {{ .Email }} {{ end }}
    ----


{{ end }}
"