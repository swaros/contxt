package ctemplate_test

import (
	"strings"
	"testing"

	"github.com/swaros/contxt/module/ctemplate"
)

func TestIgnoreHndlBase(t *testing.T) {
	origin := `Hello World
we "replacing" the word World.
execpt for any masked word, that is defined before to exclude from being replaced.
so we ignore this [World] and this (World) but any other World should be replaced.
	`
	expected := `Hello Mars
we "replacing" the word Mars.
execpt for any masked word, that is defined before to exclude from being replaced.
so we ignore this [World] and this (World) but any other World should be replaced.
	`

	ignoreHndl := ctemplate.NewIgnorreHndl(origin)
	ignoreHndl.AddIgnores("[World]", "(World)", "other World")
	maskedStr := ignoreHndl.CreateMaskedString()
	maskedStr = strings.ReplaceAll(maskedStr, "World", "Mars")

	restored := ignoreHndl.RestoreOriginalString(maskedStr)
	if restored != expected {
		t.Errorf("Expected\n%s\ngot\n%s", expected, restored)
	}

}

func TestWithGoreleaserDefaultConf(t *testing.T) {
	origin := `builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - windows
      #- darwin
    main: ./cmd/v2/main.go
    ldflags:
      - -X github.com/swaros/contxt/configure.mainversion={{.Version}}
      - -X github.com/swaros/contxt/configure.build={{.Date}}
archives:
  - format: tar.gz
    # this name template makes the OS and Arch compatible with the results of .
    name_template: >-
      {{ .ProjectName }}_
      {{- title .Os }}_
      {{- if eq .Arch "amd64" }}x86_64
      {{- else if eq .Arch "386" }}i386
      {{- else }}{{ .Arch }}{{ end }}
      {{- if .Arm }}v{{ .Arm }}{{ end }}
    # use zip for windows archives
    format_overrides:
      - goos: windows
        format: zip
        
checksum:
  name_template: 'checksums.txt'
snapshot:
  name_template: "{{ .Tag }}-next"
changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'`
	ignoresAsString := `{{ .ProjectName }}_
{{- title .Os }}_
{{- if eq .Arch "amd64" }}x86_64
{{- else if eq .Arch "386" }}i386
{{- else }}{{ .Arch }}{{ end }}
{{- if .Arm }}v{{ .Arm }}{{ end }}
{{.Date}}
{{.Version}}
{{ .Os }}
{{ .Arch }}
{{ .Arm }}
{{ .Tag }}`
	ignoreHndl := ctemplate.NewIgnorreHndl(origin)
	ignoreHndl.AddIgnores(strings.Split(ignoresAsString, "\n")...)
	maskedStr := ignoreHndl.CreateMaskedString()
	if maskedStr == origin {
		t.Error("No masking done. we have the same string")
	}
	// check if the masked string contains any of the ignored strings
	for _, ignore := range strings.Split(ignoresAsString, "\n") {
		if strings.Contains(maskedStr, ignore) {
			t.Errorf("Masked string contains ignored string %s", ignore)
		}
	}

	// check if we found any markups like {{ or }} in the masked string
	if strings.Contains(maskedStr, "{{") || strings.Contains(maskedStr, "}}") {
		t.Error("Masked string contains markup")
		// go through the string and check if we have any markup
		for no, line := range strings.Split(maskedStr, "\n") {
			if strings.Contains(line, "{{") || strings.Contains(line, "}}") {
				t.Errorf("Masked string contains markup in line:%d:\n %s", no, line)
			}
		}
	}

	restored := ignoreHndl.RestoreOriginalString(maskedStr)
	if restored != origin {
		t.Errorf("Expected\n%s\ngot\n%s", origin, restored)
	}

}
