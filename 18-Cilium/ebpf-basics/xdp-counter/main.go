package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
)

func main() {
	// Allow the current process to lock memory for eBPF resources.
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatal(err)
	}

	// Load pre-compiled programs and maps into the kernel.
	var objs counterObjects
	if err := loadCounterObjects(&objs, nil); err != nil {
		log.Fatalf("loading objects: %v", err)
	}
	defer objs.Close()

	// Get the loopback network interface
	ifaceName := "lo" // Change to eth0 or your active interface to test real traffic
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		log.Fatalf("getting interface %s: %v", ifaceName, err)
	}

	// Attach the program to the interface
	l, err := link.AttachXDP(link.XDPOptions{
		Program:   objs.CountPackets,
		Interface: iface.Index,
	})
	if err != nil {
		log.Fatalf("attaching XDP: %v", err)
	}
	defer l.Close()

	fmt.Printf("Counting packets on %s... Press Ctrl+C to stop.\n", ifaceName)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			var count uint64
			key := uint32(0)
			if err := objs.PktCount.Lookup(&key, &count); err != nil {
				log.Printf("reading map: %v", err)
				continue
			}
			fmt.Printf("\rPackets: %d", count)
		case <-stop:
			fmt.Println("\nDetaching XDP program...")
			return
		}
	}
}
