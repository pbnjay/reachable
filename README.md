# reachable [![GoDoc](https://godoc.org/github.com/pbnjay/reachable?status.svg)](http://godoc.org/github.com/pbnjay/reachable)
Network reachability functions for client side Go network code


Package reachable adds utility methods to notify when a network connection
to a remote host is up or down. This is intended for long-running programs
on systems that might have intermittent network access (e.g. laptops, phones,
remote embedded systems, etc).

## Quick Start

```go
    // check if google.com is reachable every 5 minutes
    reachable.DefaultInterval = time.Minute*5
    reachable.Start("google.com")
    defer reachable.Stop()

    // ...

    // about to do network stuff...
    if !reachable.NetworkIsReachable() {
        log.Println("no network available!")
    }
```

## Monitoring multiple hosts

This package may also be used to monitor multiple hosts by setting up
separate Checker instances:

```go
    // these will be updated whenever you need them later
    googleIsUp := true
    bingIsUp := true

    c1 := Checker{
        Hostport: "google.com:443",
        Notifier: func(r bool) {
           googleIsUp = r
        },
    }

    c2 := Checker{
        Hostport: "bing.com",
        Notifier: func(r bool) {
           bingIsUp = r
        },
    }

    // start goroutines that check for reachability
    c1.Start()
    c2.Start()

    // ...
```

## License

MIT