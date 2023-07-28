# mimicry log

this module exists, because I wanted to have a basic abstraction of logging without any Implementation.
> there might be similar modules out there, but I just don't find them. at least none that was matching my needs.

#### the idea

while moving some logic from the main project to independent modules, I wanted to avoid forcing a specific logger implementation. like `logrus` or `zap`. the main app was using `logrus` and the first step while moving to modules was just to remove any dependecies to `logrus` and decide later which logger to use.

all of them could also be done by just using the `log` package, but I wanted to have a bit more control over the log level and the output format. and also I wanted to have a bit more control over the logger implementation.

this idea is not a new one, and it exists in many other languages, and also in go. but what i found so far was different in "mindset" and in some points did not match my needs.

### the needs

- there should be no fancy stuff in there, and no new concepts.
- no predefined Logger specific arguments, like `logrus.Fields` or `zap.Field`
- no predefined Logger specific return values, like `logrus.Entry` or `zap.SugaredLogger`
- no predefined Logger specific methods, like `logrus.WithField` or `zap.Sugar`
- the implementation should be easy to replace
- the implementation have to take care about any specific Logger arguments, return values and methods without exposing them to the user
- the implementation have to handle all arguments and map them to the specific Logger arguments, return values and methods

## the interface

### logger methods

these methods are the basic logging methods. they are implemented by the specific logger implementation.

```go
    Trace(args ...interface{})
    Debug(args ...interface{})
    Info(args ...interface{})
    Error(args ...interface{})
    Warn(args ...interface{})
    Critical(args ...interface{})
```

### logger level checks
these methods are used to check if a specific level is enabled. 
```go
    IsLevelEnabled(level string) bool 
    IsTraceEnabled() bool
    IsDebugEnabled() bool
    IsInfoEnabled() bool
    IsWarnEnabled() bool
    IsErrorEnabled() bool
    IsCriticalEnabled() bool
```

### logger level set/get methods
these methods are used to set and get the log level. the level is a string, and the implementation have to map it to the specific logger level.
```go
    SetLevelByString(level string) 
    SetLevel(level interface{})
    GetLevel() string
```
### logger getter
this method is used to get the specific logger implementation. the implementation have to return the specific logger implementation.
```go
    GetLogger() interface{}    
```

## Implemented
mimiclog is mainly a interface. there is only one implementation, and that is the `NullLogger` implementation. This is a logger that does nothing. it is used as default logger, and can be used as a fallback logger.

### Apply helper function
a second interface defines how to apply a logger to a struct.
the only needed method is `SetLogger(logger mimiclog.Logger)`.

```go
    type MimicLogUser interface {
        SetLogger(logger Logger)
    }
```
so then it is easy to apply the logger.
```go
    logger := NewLogrusLogger()
    app := &myApp{}
    mimiclog.ApplyLogger(logger, app)
```
this can be also done by try and error. (*for example some dynamic dependencies*)
```go
    logger := NewLogrusLogger()
    theLib := &someLib{}
    if ok := mimiclog.ApplyIfPossible(logger, theLib); !ok {
        logger.Warn(
            "'theLib' seems not supporting the 'mimiclog' interface"
        )
    }
```