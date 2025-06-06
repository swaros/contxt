### This file is generated by using ./imports/go_build.tpl
### Any change you are doing here is overwritten while the next contxt build
### change go_build.tpl instead
name: Go

on:
  push:
    branches: [ main  ]
  pull_request:
    branches: [ wip/release** ]

jobs:

  build:
    name: Build
    runs-on: ubuntu-latest
    steps:

    - name: Set up Go 1.x
      uses: actions/setup-go@v5
      with:
        go-version: ^1.18
      id: go

    - name: Check out code into the Go module directory
      uses: actions/checkout@v4

{{- range $targetName, $targets := $.build.targets}}
  {{- if $targets.is_release}}
    - name: Build-{{ $targetName }}
      run: go build -ldflags "{{- range $k, $ldflag := $.build.preset.ldflags }} -X {{ $ldflag }} {{- end -}} {{- range $k, $ldflag := $targets.ldflags }} -X {{ $ldflag }} {{- end -}}" -o ./bin/{{ $targets.output }} {{ $targets.mainfile }}
 {{- end -}}
{{- end }}

    - name: Test
      run: |
    {{- range $k, $test := $.module }}
    {{- if $test.local}}
          go test  -failfast ./module/{{ $test.modul }}/./...
    {{- end }}
    {{- end }}

#{{- range $targetName, $targets := $.build.targets}}
#  {{- if $targets.is_release}}
#    - name: contxt-artifact-{{ $targetName }}
#      uses: actions/upload-artifact@v4
#      with:
#        name: "contxt-{{ $targetName }}"
#        path: ./build/{{ $targets.output }}

# {{- end -}}
#{{- end }}
