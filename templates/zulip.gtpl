# {{ if .Report.Escalated }}:fire:{{ end }}{{ .Report.Severity | severityToEmoji }} {{ .Report.Severity.String | upper }}: {{ .Report.Subject }}
## {{ .Report | mkReportSourceIDLabel }}

{{ .Report.Body }}
{{ if and .Info.PublicPath .Report.ID }}
```spoiler Report actions
[Inspect report]({{ .Info.PublicPath }}/ui/detail?id={{ .Report.ID }})
[List group]({{ .Info.PublicPath }}/ui/list?app={{ .Report.SourceID.App }}&instance={{ .Report.SourceID.Instance }}&tag={{ .Report.SourceID.Tag }})
```
{{ end }}

*Generated by {{ .NotifierName }}/Conomi{{ if .Info.Build.Version }} {{ .Info.Build.Version }}{{ end }}*