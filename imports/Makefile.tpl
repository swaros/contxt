build: test	build-no-test

build-no-test:
	go build -ldflags "-X ${configmodul}.minversion=${minor-version} -X ${configmodul}.midversion=${mid-version} -X ${configmodul}.mainversion=${main-version} -X ${configmodul}.build=${build-hash}" -o ./bin/contxt cmd/cmd-contxt/main.go

install-local: build
	./bin/contxt run install-local

install-local-no-test: build-no-test
	./bin/contxt run install-local

clean:
	rm -f ./bin/contxt
	rm -rf ./dist

test:
{{- range $k, $test := $.module }}
  {{- if $test.local }}
	go test  -failfast ./module/{{ $test.modul }}/./...
  {{- end }}
{{- end }}

info: build
	./bin/contxt dir