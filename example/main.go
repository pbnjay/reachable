package main

import (
	"flag"
	"os"
	"time"

	"github.com/pbnjay/reachable"
)

func main() {
	host := flag.String("h", "google.com", "hostname to check")
	ival := flag.Duration("i", time.Second*5, "interval for reachability checks")
	flag.Parse()

	reachable.DefaultInterval = *ival
	reachable.Start(*host)
	defer reachable.Stop()

	tick := time.NewTicker(time.Second)
	for {
		<-tick.C
		if reachable.NetworkIsReachable {
			os.Stdout.Write([]byte("\r  REACHABLE  "))
		} else {
			os.Stdout.Write([]byte("\rNOT REACHABLE"))
		}
		os.Stdout.Sync()
	}
}
