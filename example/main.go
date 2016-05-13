package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/pbnjay/reachable"
)

func main() {
	host := flag.String("h", "google.com", "hostname to check")
	host2 := flag.String("h2", "bing.com", "second hostname to check")
	ival := flag.Duration("i", time.Second*5, "interval for reachability checks")
	flag.Parse()

	reachable.DefaultInterval = *ival

	// example usage for 2 domains and separate instances
	os.Stdout.Write([]byte("Toggle network a few times to see notifications:\n"))

	c1 := reachable.Checker{
		Hostport: *host,
		Notifier: func(r bool) {
			if r {
				log.Println(*host + " is UP")
			} else {
				log.Println(*host + " is DOWN")
			}
		},
	}
	c2 := reachable.Checker{
		Hostport: *host2,
		Notifier: func(r bool) {
			if r {
				log.Println(*host2 + " is UP")
			} else {
				log.Println(*host2 + " is DOWN")
			}
		},
	}

	// start goroutines that check for reachability
	c1.Start()
	c2.Start()

	// kill the server/unplug the cord/disable wifi to see notifications!
	time.Sleep(time.Minute)

	c1.Stop()
	c2.Stop()

	// example usage with the package-level functions
	os.Stdout.Write([]byte("Toggle network a few times to see live updates:\n"))

	reachable.Start(*host)
	defer reachable.Stop()

	// kill the server/unplug the cord/disable wifi again to see live updates!
	for i := 0; i < 60; i++ {
		time.Sleep(time.Second)
		if reachable.NetworkIsReachable {
			os.Stdout.Write([]byte("\r" + *host + "   REACHABLE  "))
		} else {
			os.Stdout.Write([]byte("\r" + *host + " NOT REACHABLE"))
		}
		os.Stdout.Sync()
	}
}
