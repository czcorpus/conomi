{{ .Report.Body }}<br/>
<br/>
<a href="{{ .Info.PublicPath }}/report/{{ .Report.ID }}">Inspect report</a><br/>
<a href="{{ .Info.PublicPath }}/resolve/{{ .Report.ID }}?user_id=1">Resolve report</a><br/>
<br/>
<i>From notifer {{ .NotifierName }}/Conomi{{ if .Info.Build.Version }} {{ .Info.Build.Version }}{{ end }}</i>