package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
)

func main() {
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatal(err)
	}

	var objs openatObjects
	if err := loadOpenatObjects(&objs, nil); err != nil {
		log.Fatalf("loading objects: %v", err)
	}
	defer objs.Close()

	tp, err := link.Tracepoint("syscalls", "sys_enter_openat", objs.TraceOpenat, nil)
	if err != nil {
		log.Fatalf("opening tracepoint: %s", err)
	}
	defer tp.Close()

	fmt.Println("Tracing openat syscalls... Press Ctrl+C to stop.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Println("\n--- PID openat() Call Counts ---")
			var pid uint32
			var count uint64
			iter := objs.OpenatCounts.Iterate()
			for iter.Next(&pid, &count) {
				fmt.Printf("PID: %d | Calls: %d\n", pid, count)
			}
		case <-stop:
			fmt.Println("\nDetaching tracepoint...")
			return
		}
	}
}
