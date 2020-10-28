
{{ range $k, $v := $.msgs }}Key:{{ $k }}, Value:{{ $v }}
{{ end }}
{{ range $_, $v := $.msgNum }}Values: {{ $v }}
{{ end }}
{{ $.nested }}
{{ range $_, $v := $.nested }}
	{{ if isInt $v }}
	v is int .. {{ $v }}
	{{ end }}
	{{- if isMap $v -}}
	{{- range $k, $v := $v -}}
		k={{ $k }}, v={{ $v }}
	{{- end -}}
	{{- end -}}
	{{- if isSlice $v -}}
		{{ range $_, $s := $v -}}
			{{ $s }}
		{{- end }}
	{{- end -}}
{{ end }}