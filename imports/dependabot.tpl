# To get started with Dependabot version updates, you'll need to specify which
# package ecosystems to update and where the package manifests are located.
# Please see the documentation for all configuration options:
# https://docs.github.com/github/administering-a-repository/configuration-options-for-dependency-updates
version: 2
updates:
{{- range $k, $deb := $.module }}
{{- if $deb.local }}
### dependencie for /module/{{ $deb.modul }}/
  - package-ecosystem: gomod
    directory: /module/{{ $deb.modul }}/
    schedule:
       interval: daily
    assignees:
       - swaros
{{- end }}
{{- end }}