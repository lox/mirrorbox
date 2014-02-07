Mirrorbox
=========

A local http server that servers the ubuntu mirrors.txt file with crappy mirrors removed.

Installing and Running
----------------------

Requires geoip installed.

```bash
go build github.com/lox/mirrorbox
```

Development
-----------

The gox tool is used for cross-compiling.

```bash
go get
go get github.com/mitchellh/gox
gox -build-toolchain
make 
```