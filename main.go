package main

import (
	"fmt"
	"k8s.io/klog/v2"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func run(stop <-chan struct{}) {
	<-stop
	time.Sleep(time.Second * 5)
	fmt.Println(2222)

}

func main() {
	stop := make(chan struct{})
	// listen for interrupts or the Linux SIGTERM signal and cancel
	// our context, which the leader election code will observe and
	// step down
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		klog.Info("Received termination, signaling shutdown")
		close(stop)
	}()

	run(stop)
	fmt.Println("done")
}
