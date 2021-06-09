# logos

***Logging + Printing + Compromising***

logos is a small and super-opinionated wrapper around [`go.uber.org/zap`](https://pkg.go.dev/go.uber.org/zap?utm_source=godoc) to:

- give users decent looking output
- give auditors structured logs to analyze
- give developers convenient functions to call both

With default settings, logos user output looks like:

```
INFO: Now we're logging :)
  key: "value"
  otherkey: "othervalue"
```

Assuming `NewZapSugaredLogger` is used to create the logger,
the same `logger.Infow` looks like the following in the log.

```
{"_level":"INFO","_timestamp":"2021-06-08T22:16:29.161-0700","_caller":"logos/example_logos_test.go:21","_function":"github.com/bbkane/logos_test.Example","_msg":"Now we're logging :)","_pid":49721,"_version":"v1.0.0","key":"value","otherkey":"othervalue"}
```

Note that logos can wrap any `zap.Logger`, the provided `NewZapSugaredLogger` is only a convenience function.

## Use

logos *might* be useful for small CLI apps that need logs

logos *won't* be useful for performance sensitive apps (no attention paid to allocation, unbuffered prints), apps designed to produce output for piping to another command, or apps producing deeply nested logs.

logos is imported by [these packages](https://pkg.go.dev/github.com/bbkane/sugarkane?tab=importedby).

See the [pkg.go.dev docs](https://pkg.go.dev/github.com/bbkane/logos) for the exact API and example usage.

## Philosophy

`logos` users (i.e., the author 😃) only believe in 3 log levels: 

- `DEBUG` for information only needed by auditors looking at the logs
- `ERROR` for problems
- `INFO` for other information

Correspondingly, it offers the following functions and destinations for their content:

|               | stderr | stdout | logfile |
| ------------- | :----: | :----: | :-----: |
| Errorw        |   x    |        |         |
| Infow         |        |   x    |         |
| Logger.Debugw |        |        |    x    |
| Logger.Errorw |   x    |        |    x    |
| Logger.Infow  |        |   x    |    x    |

The logger functions are a subset of `zap.SugaredLogger` so if your app gets too large, you can do a bit of work and swap them.

In addition, logos offers `Logger.Sync` to sync the logs and `Logger.LogOnPanic` as an optional function to `recover` from a panic, log, and then panic again.

## Analyzing JSON Logs

TODO

## History

logos began as a set of functions in [`grabbit`](https://github.com/bbkane/grabbit), so I could have a log of failed image downloads to analyze. Eventually I extracted it into [`sugarkane`](https://github.com/bbkane/sugarkane) to use in other apps. Finally, I reworked the API and released logos.
