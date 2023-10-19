echo "build script"
{{- range $targetName, $targets := $.build.targets}}
  {{- if $targets.is_release}}

### build {{ $targetName }}

# compose build ldflags
BUILD_ARGS="{{- range $k, $ldflag := $.build.preset.ldflags }} -X {{ $ldflag }} {{- end -}} {{- range $k, $ldflag := $targets.ldflags }} -X {{ $ldflag }} {{- end -}}"
# build {{ $targetName }}
echo "-> build {{ $targetName }}"
go build -ldflags "$BUILD_ARGS" -o ./bin/{{ $targets.output }} {{ $targets.mainfile }}
./bin/{{- $targets.output }} {{ $targets.version_verify }}
echo "-> build {{ $targetName }} done"

 
 {{- end -}}
{{- end -}}
