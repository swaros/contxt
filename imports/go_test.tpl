name: wip-test
on:
  push:
    branches:
      - "wip/**"

jobs:

  build:
    name: Build
    runs-on: ubuntu-latest
    steps:

    - name: Set up Go 1.x
      uses: actions/setup-go@v2
      with:
        go-version: ^1.18
      id: go

    - name: Check out code into the Go module directory
      uses: actions/checkout@v2

    - name: Build contxt bin
      run: go build -ldflags "-X github.com/swaros/contxt/configure.minversion={{ $.release.version.minor }} -X github.com/swaros/contxt/configure.midversion={{ $.release.version.mid }} -X github.com/swaros/contxt/configure.mainversion={{ $.release.version.main }} -X github.com/swaros/contxt/configure.build=`date -u +.%Y%m%d.%H%M%S`" -o ./bin/contxt cmd/cmd-contxt/main.go

    - name: Test
      run: |
          ./bin/contxt run test-each
