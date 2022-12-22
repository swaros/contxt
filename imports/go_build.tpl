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
      uses: actions/setup-go@v2
      with:
        go-version: ^1.18
      id: go

    - name: Check out code into the Go module directory
      uses: actions/checkout@v2

    - name: Build
      run: go build -ldflags "-X github.com/swaros/contxt/configure.minversion={{ $.release.version.minor }} -X github.com/swaros/contxt/configure.midversion={{ $.release.version.mid }} -X github.com/swaros/contxt/configure.mainversion={{ $.release.version.main }} -X github.com/swaros/contxt/configure.build=`date -u +.%Y%m%d.%H%M%S`" -o ./bin/contxt cmd/cmd-contxt/main.go

    - name: Test
      run: |
        ./bin/contxt run test-each        
    
    - name: contxt-artifact
      uses: actions/upload-artifact@v2
      with:
        name: "contxt-bin"
        path: ./bin/contxt
