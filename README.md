# tpl

[![build](https://github.com/zooplus/tpl/actions/workflows/build-go.yml/badge.svg?branch=main)](https://github.com/zooplus/tpl/actions/workflows/build-go.yml)
![GitHub Release Date](https://img.shields.io/github/release-date/zooplus/tpl)

`tpl` is build for generating config files from templates using simple or complex (lists, maps, objects) shell environment
variables. Since the binary has zero dependencies it is build for Docker but you can use it across all platform and
operating systems.

`tpl` uses [sprig](https://github.com/Masterminds/sprig) to extend golang's template capabilities.

Check the test section and have a look at [`test/test.tpl`](test/test.tpl) (template) and [`test/test.txt`](test/test.txt) (result) in `test` folder for examples.

## setup

Just download the binary for your OS and arch from the [releases](https://github.com/zooplus/tpl/releases) page.

If you want to use it inside your docker image you can add this to your `Dockerfile`:

```
ADD https://github.com/zooplus/tpl/releases/download/v0.12.4/tpl-linux-amd64 /usr/local/bin/tpl
RUN chmod a+x /usr/local/bin/tpl
```

## build

Local:
```
go install github.com/zooplus/tpl@latest
```

Cross platform builds happen on dedicated runners on GitHub Actions.
```
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0    # for static binary
go build \
    -o bin/tpl \
    .
```

## unit test

```
go test -v
```

## integration tests

```bash
./test.sh
```
