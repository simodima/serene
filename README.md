# Grace

Package `github.com/simodima/skeleton/bones`
A graceful shutdown HTTP server


```go
err := grace.RunGracefully(
    router,
    grace.WithBindAddress(":8080"),
    grace.WithShutdownTimeout(5*time.Second),
)
```

## Healthz handler
Package `github.com/simodima/skeleton/healthz`

```go
healthz.HealthzHandler(router,func() (healthz.Dependency, bool) {
    // check your dependency
    // isDependencyOK := true
    // return healthz.Dependency{..}, isDependencyOK
})
```