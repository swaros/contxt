echo "build script"
{{- range $targetName, $targets := $.build.targets}}
  {{- if $targets.is_release}}
echo "-> build {{ $targetName }}"
go build -ldflags "{{- range $k, $ldflag := $.build.preset.ldflags }} -X {{ $ldflag }} {{- end -}} {{- range $k, $ldflag := $targets.ldflags }} -X {{ $ldflag }} {{- end -}}" -o ./bin/{{ $targets.output }}.exe {{ $targets.mainfile }}
echo "-> build {{ $targetName }} done"
 
 {{- end -}}
{{- end -}}
