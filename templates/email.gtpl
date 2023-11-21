{{ .Report.Body }}<br/>
<br/>
{{ if and .Info.PublicPath .Report.ID }}
<a href="{{ .Info.PublicPath }}/ui/detail?id={{ .Report.ID }}">Inspect report</a><br/>
<a href="{{ .Info.PublicPath }}/ui/list?app={{ .Report.SourceID.App }}&instance={{ .Report.SourceID.Instance }}&tag={{ .Report.SourceID.Tag }}">List group</a><br/>
<br/>
{{ end }}
<i>From notifer {{ .NotifierName }}/Conomi{{ if .Info.Build.Version }} {{ .Info.Build.Version }}{{ end }}</i>