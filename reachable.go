// Package reachable adds utility methods to be notifies if a network connection
// goes up or is down.
package reachable

import (
	"net"
	"time"
)

var (
	// NetworkIsReachable is set to true when the package is able to reach the
	// configured host. Start("example.com") must be call for this to be valid.
	// If Start is not called, the default value ensures code continues to work.
	NetworkIsReachable = true

	// DefaultInterval is the polling interval when network checks are made.
	DefaultInterval = time.Minute

	singleton = &Checker{}
)

// Checker is a reachability checker that can notify calling code when network
// access comes up or goes away.
type Checker struct {
	// Hostname to contact to verify connectivity.
	Hostname string
	// Port to use (including ":" prefix) during TCP connection (default ":80").
	Port string
	// Interval to poll for network access.
	Interval time.Duration
	// Notifier is the user-specified callback for reachability notifications.
	Notifier func(bool)

	quit chan struct{}
}

// Start begins Checker polling in a background goroutine.
func (c *Checker) Start() {
	if c.Port == "" {
		c.Port = ":80"
	}
	c.quit = make(chan struct{})
	go singleton.run()
}

// Stop tells the background goroutine to stop checking.
func (c *Checker) Stop() {
	c.quit <- struct{}{}
}

// Start begins the default Checker instance with the DefaultInterval and
// updates the global NetworkIsReachable boolean value.
func Start(hostname string) {
	singleton.Hostname = hostname
	singleton.Port = ":80"
	h, p, err := net.SplitHostPort(hostname)
	if err == nil {
		singleton.Hostname = h
		singleton.Port = ":" + p
	}
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
	conn, err := net.DialTimeout("tcp", c.Hostname+c.Port, time.Second*3)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func (c *Checker) run() {
	currentStatus := -1
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
