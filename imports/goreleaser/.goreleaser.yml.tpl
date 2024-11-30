# This is an example goreleaser.yaml file with some sane defaults.
# Make sure to check the documentation at http://goreleaser.com
# yaml-language-server: $schema=https://goreleaser.com/static/schema.json
# vim: set ts=2 sw=2 tw=0 fo=cnqoj
version: 2
before:
  hooks:
    # You may remove this if you don't use go modules.
    #- go mod download
    # you may remove this if you don't need go generate
    # - go generate ./...
builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - windows
      #- darwin
    main: ./cmd/v2/main.go
    ldflags:
      {{- range $k, $ldflag := $.build.preset.ldflags }}
      - -X {{ $ldflag }} 
      {{- end }} 
      {{- range $k, $ldflag := $.release.ldflags }}
      - -X {{ $ldflag }} 
      {{- end }}
archives:
  - format: tar.gz
    # this name template makes the OS and Arch compatible with the results of `uname`.
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
  version_template: "{{ .Tag }}-next"
changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'

env_files:
  github_token: ~/keystore/tokens/goreleaser.token

nfpms:
 -
    id: contxt-rpm
    vendor: swaros
    homepage: https://github.com/swaros/contxt
    maintainer: tziegler <thomas.zglr@googlemail.com>
    description: managing paths and jobs in workspaces on console.
    license: MIT
    formats:
      - rpm
      - deb
      - apk

    rpm:
      summary: context manager
      compression: lzma

