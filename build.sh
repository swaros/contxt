echo "build script"

### build release

# compose build ldflags
BUILD_ARGS=" -X github.com/swaros/contxt/module/configure.minversion=3 -X github.com/swaros/contxt/module/configure.midversion=5 -X github.com/swaros/contxt/module/configure.mainversion=0 -X github.com/swaros/contxt/module/configure.build=.20240808.142542-linux-release"
# build release
echo "-> build release"
go build -ldflags "$BUILD_ARGS" -o ./bin/contxt cmd/cmd-contxt/main.go
./bin/contxt version
echo "-> build release done"