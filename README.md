# gitempl

Provides a means to generated templated arbitrary text from
a git repo. The generator expects conventional commits. The
template provides should be a normal Go template. You'll have
the following context available when the template executes:

Higher level construct Commits is provided, a slice of commits
with conventional commit information baked in at the .CC field
of each commit.

You can run the following with the template below from the root of a git repo:

```shell
gitempl <<EOF
{{ range .Commits }}
    {{ .Author }} commited by
    {{ .Hash }} commit hash
    {{ .HashShort }} commmit hash truncated to 7 chars
    {{ .Message }} commit message full
    {{ .Stats }} commit stats (additions/removals)
    
    {{ with .CC }}
        Abbreviated access using to conventional commit (CC) data:
        {{ .Body }} body of CC
        {{ .Desc }} description of CC
        {{ .Footer }} footer of CC
        {{ .Header }} header of CC
        {{ .Scope }} scope of CC
        {{ .Type }} type of CC
        {{ range .Notes }}
          {{ .Type }} type of note
          {{ .Value }} value of note
        {{ end }}
    {{ end }}

{{ end }}
EOF
```

For more information, see the `gitempl -h` usage.