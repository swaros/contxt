build: test	build-no-test

build-no-test:
{{- range $targetName, $targets := $.build.targets}}
{{- if $targets.is_release}}
	go build -ldflags "{{- range $k, $ldflag := $.build.preset.ldflags }} -X {{ $ldflag }} {{- end -}} {{- range $k, $ldflag := $targets.ldflags }} -X {{ $ldflag }} {{- end -}}" -o ./bin/{{ $targets.output }} {{ $targets.mainfile }}
{{- end }}
{{- end }}

{{- range $targetName, $targets := $.build.targets}}
build-{{ $targetName }}:
	go build -ldflags "{{- range $k, $ldflag := $.build.preset.ldflags }} -X {{ $ldflag }} {{- end -}} {{- range $k, $ldflag := $targets.ldflags }} -X {{ $ldflag }} {{- end -}}" -o ./bin/{{ $targets.output }} {{ $targets.mainfile }}
{{- end }}

clean:
	rm -f ./bin/contxt
	rm -rf ./dist

test:
{{- range $k, $test := $.module }}
  {{- if $test.local }}
	go test  -failfast ./module/{{ $test.modul }}/./...
  {{- end }}
{{- end }}
{{- if $.testcases.runner}}
	{{ $.testcases.runner }} run test-loops
{{- end }}

info: build
	./bin/contxt dir