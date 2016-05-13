// Package reachable adds utility methods to notify when a network connection
// to a remote host is up or down. This is intended for long-running programs
// on systems that might have intermittent network access (e.g. laptops, phones,
// remote embedded systems, etc).
//
//    // check if google.com is reachable every 5 minutes
//    reachable.DefaultInterval = time.Minute*5
//    reachable.Start("google.com")
//    defer reachable.Stop()
//
//    ...
//
//    // about to do network stuff...
//    if !reachable.NetworkIsReachable {
//      log.Println("no network available!")
//    }
//
// This package may also be used to monitor multiple hosts by setting up
// separate Checker instances:
//
//    // these will be updated whenever you need them
//    googleIsUp := true
//    bingIsUp := true
//
//    c1 := Checker{
//        Hostport:"google.com:443",
//        Notifier: func(r bool) {
//           googleIsUp = r
//        },
//    }
//    c2 := Checker{
//        Hostport:"bing.com",
//        Notifier: func(r bool) {
//           bingIsUp = r
//        },
//    }
//
//    // start goroutines that check for reachability
//    c1.Start()
//    c2.Start()
//
//    ...
//
package reachable

import (
	"net"
	"strings"
	"time"
)

var (
	// NetworkIsReachable is set to true when the package is able to reach the
	// configured host. Start("example.com") must be called for this to be valid.
	// If Start is not called, the default value ensures code continues to work.
	NetworkIsReachable = true

	// DefaultInterval is the polling interval when network checks are made.
	DefaultInterval = time.Minute

	// DefaultTimeout specifies how long the TCP connection attempt should wait
	// before timing out. This should be adjusted for high-latency connections.
	DefaultTimeout = time.Second * 3

	singleton = &Checker{}
)

// Checker is a reachability checker that notifies calling code when a given
// host and port is reachable via the network.
type Checker struct {
	// Hostport contains the hostname and port to contact to verify
	// connectivity. If no port is provided, assumes default port 80.
	Hostport string

	// Interval to poll for network access. If zero or negative, uses DefaultInterval.
	Interval time.Duration

	// Notifier is the user-specified callback for reachability notifications.
	Notifier func(bool)

	quit chan struct{}
}

// Start begins Checker polling in a background goroutine.
func (c *Checker) Start() {
	c.quit = make(chan struct{})
	go c.run()
}

// Stop tells the background goroutine to stop checking.
func (c *Checker) Stop() {
	c.quit <- struct{}{}
}

// Start begins the default Checker instance with the DefaultInterval and
// updates the global NetworkIsReachable boolean value.
func Start(hostname string) {
	singleton.Hostport = hostname
	singleton.Interval = DefaultInterval
	singleton.Notifier = func(a bool) {
		NetworkIsReachable = a
	}
	singleton.Start()
}

// Stop the global instance and reset NetworkIsReachable to true.
func Stop() {
	singleton.Stop()
	// keep system in a sane/useful state when not running
	NetworkIsReachable = true
}

func (c *Checker) hasInterfaceUp() bool {
	ifaces, err := net.Interfaces()
	if err != nil {
		return false
	}
	for _, x := range ifaces {
		if (x.Flags & net.FlagLoopback) != 0 {
			// loopback doesn't help
			continue
		}
		if (x.Flags & net.FlagUp) != 0 {
			return true
		}
	}
	return false
}

func (c *Checker) canConnect() bool {
	if !strings.Contains(c.Hostport, ":") {
		c.Hostport += ":80"
	}
	conn, err := net.DialTimeout("tcp", c.Hostport, DefaultTimeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func (c *Checker) run() {
	currentStatus := -1
	if c.Interval <= time.Duration(0) {
		c.Interval = DefaultInterval
	}
	t := time.NewTicker(c.Interval)
	for {
		select {
		case <-c.quit:
			t.Stop()
			close(c.quit)
			return

		case <-t.C:
			isActive := c.hasInterfaceUp()
			if isActive {
				isActive = c.canConnect()
			}
			if !isActive {
				if currentStatus != 0 { // already inactive?
					c.Notifier(false)
					currentStatus = 0
				}
			} else {
				if currentStatus != 1 { // already active?
					c.Notifier(true)
					currentStatus = 1
				}
			}
		}
	}
}
