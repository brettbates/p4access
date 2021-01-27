action: RESPOND
message:  "
{{ range $group := .Groups }}
    Group {{ $group.Group }} grants {{ $group.Access }} access to the path: 

        {{ $group.Path }}

    You can get access by contacting one of the owners listed: 
    {{ range $group.Owners }} 
        {{ . }} {{ end }}
{{ end }}
"