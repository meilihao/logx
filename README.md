# logx

logx is a go logs manager, with level, rotate, adapter. it changed for [beego/logs](github.com/astaxie/beego/logs).

## What adapters are supported?

As of now logx support console, file ,multifile.

## How to use it?

see *_test.go.

### console

```go
log := NewLogger()
log.AddLogger("console", `{"color":false}`)  
```

### file

```go
log := NewLogger()
log.AddLogger("file", `{"filename":"app.log","maxlines":0,"maxsize":0,"daily":true,"maxdays":10,"perm": "0666"}`)
```

### multifile

```go
log := NewLogger()
log.AddLogger("multifile", `{"filename":"app.log","maxlines":0,"maxsize":0,"daily":true,"maxdays":10,"perm": "0666","separate":["debug", "info"]}`)
```

## 改进

1. 弃用`Register`机制

由于采用`Register`机制,`adapter.Init()`时是用新配置里的同名参数覆盖旧配置,而未指定的参数仍沿用旧参数,导致logx的行为与期望不符,比如

```go
log := NewLogger()
log.AddLogger("file", `{"filename":"test3.log","maxlines":4}`)

log2 := NewLogger()
log2.AddLogger("file", `{"filename":"test4.log"}`)
// log2 adapter里的MaxLines其实是4.
```

这种情况在运行*_test.go的时候比较常见.

## LICENSE

Apache Licence, Version 2.0 (http://www.apache.org/licenses/LICENSE-2.0.html).