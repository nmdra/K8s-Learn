# 🐝 eBPF Go Basics — Learn eBPF with Go

> Hands-on project to learn eBPF programming using the [`cilium/ebpf`](https://github.com/cilium/ebpf) Go library ([ebpf-go.dev](https://ebpf-go.dev/)).

---

## Table of Contents

- [What is eBPF?](#what-is-ebpf)
- [What is ebpf-go?](#what-is-ebpf-go)
- [Prerequisites](#prerequisites)
- [Project Setup](#project-setup)
- [Architecture Overview](#architecture-overview)
- [Getting Started](#getting-started)
  - [1. Initialize the Go Module](#1-initialize-the-go-module)
  - [2. Write the eBPF C Program](#2-write-the-ebpf-c-program)
  - [3. Generate Go Bindings](#3-generate-go-bindings)
  - [4. Write the Go Userspace Loader](#4-write-the-go-userspace-loader)
  - [5. Build & Run](#5-build--run)
- [Key Concepts](#key-concepts)
  - [eBPF Program Types](#ebpf-program-types)
  - [eBPF Maps](#ebpf-maps)
  - [Attach Points (Hooks)](#attach-points-hooks)
  - [bpf2go Code Generation](#bpf2go-code-generation)
- [Examples](#examples)
  - [Example 1: Counting Packets (XDP)](#example-1-counting-packets-xdp)
  - [Example 2: Tracing Syscalls (Tracepoint)](#example-2-tracing-syscalls-tracepoint)
  - [Example 3: Monitoring TCP Connections (Kprobe)](#example-3-monitoring-tcp-connections-kprobe)
- [Debugging & Troubleshooting](#debugging--troubleshooting)
- [Resources](#resources)

---

## What is eBPF?

**eBPF (extended Berkeley Packet Filter)** is a technology that allows you to run sandboxed programs inside the Linux kernel without modifying kernel source code or loading kernel modules. It's used for:

- 🔍 **Observability** — Tracing, profiling, monitoring
- 🛡️ **Security** — Runtime enforcement, syscall filtering
- 🌐 **Networking** — High-performance packet processing (XDP, TC)

```
┌─────────────────────────────────────┐
│           User Space                │
│  ┌──────────────────────────────┐   │
│  │  Go Application (ebpf-go)   │   │
│  │  - Load eBPF bytecode       │   │
│  │  - Read eBPF maps           │   │
│  │  - Attach to hooks          │   │
│  └──────────┬───────────────────┘   │
│             │ bpf() syscall         │
├─────────────▼───────────────────────┤
│           Kernel Space              │
│  ┌──────────────────────────────┐   │
│  │  eBPF Verifier               │   │
│  │  (safety checks)             │   │
│  └──────────┬───────────────────┘   │
│             │                       │
│  ┌──────────▼───────────────────┐   │
│  │  eBPF Programs (JIT compiled)│   │
│  │  - XDP / TC / Tracepoints   │   │
│  │  - Kprobes / Fentry          │   │
│  └──────────┬───────────────────┘   │
│             │                       │
│  ┌──────────▼───────────────────┐   │
│  │  eBPF Maps (shared state)    │   │
│  │  - Hash / Array / RingBuf    │   │
│  └──────────────────────────────┘   │
└─────────────────────────────────────┘
```

---

## What is ebpf-go?

[`ebpf-go`](https://ebpf-go.dev/) is a **pure-Go library** to read, modify, and load eBPF programs and attach them to various hooks in the Linux kernel. It's maintained by the **Cilium community** and is the foundation of Cilium's eBPF infrastructure.

**Why ebpf-go?**
- No CGo dependency — pure Go
- Compile eBPF C code at build time via `bpf2go`
- Type-safe Go bindings for eBPF maps and programs
- Production-proven (powers Cilium, Tetragon, Cloudflare, etc.)

---

## Prerequisites

- **Linux** kernel ≥ 5.4 (5.10+ recommended)
- **Go** ≥ 1.21
- **Clang/LLVM** ≥ 14 (for compiling eBPF C programs)
- **libbpf headers** (usually installed with `libbpf-dev` or similar)
- **Linux headers** for your kernel version

### Install Dependencies (Ubuntu/Debian)

```bash
sudo apt-get update
sudo apt-get install -y clang llvm libbpf-dev linux-headers-$(uname -r) gcc-multilib
```

### Install Dependencies (Fedora)

```bash
sudo dnf install -y clang llvm libbpf-devel kernel-headers
```

---

## Project Setup

```
ebpf-go-basics/
├── README.md
├── go.mod
├── go.sum
├── headers/              # vmlinux.h & BPF helper headers
│   └── .gitkeep
├── xdp-counter/          # Example 1: XDP packet counter
│   ├── counter.c         # eBPF C program
│   ├── main.go           # Go loader / userspace program
│   └── gen.go            # go:generate directive for bpf2go
├── tracepoint-openat/    # Example 2: Trace openat syscalls
│   ├── openat.c
│   ├── main.go
│   └── gen.go
└── kprobe-tcp/           # Example 3: Monitor TCP connects
    ├── tcpconnect.c
    ├── main.go
    └── gen.go
```

---

## Architecture Overview

```
┌──────────────────┐    bpf2go     ┌──────────────────┐
│  counter.c       │ ──────────▶   │  counter_bpfe*.go │
│  (eBPF C code)   │  (go generate)│  counter_bpfe*.o  │
└──────────────────┘               └────────┬─────────┘
                                            │
                                   ┌────────▼─────────┐
                                   │    main.go        │
                                   │  (Go userspace)   │
                                   │  - Load .o        │
                                   │  - Attach to hook │
                                   │  - Read maps      │
                                   └──────────────────┘
```

---

## Getting Started

### 1. Initialize the Go Module

```bash
cd ebpf-go-basics
go mod init ebpf-go-basics
go get github.com/cilium/ebpf@latest
go get github.com/cilium/ebpf/cmd/bpf2go@latest
```

### 2. Write the eBPF C Program

```c
// xdp-counter/counter.c
//go:build ignore

#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

// Map to store the packet count
struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __type(key, __u32);
    __type(value, __u64);
    __uint(max_entries, 1);
} pkt_count SEC(".maps");

// XDP program that counts every packet
SEC("xdp")
int count_packets(struct xdp_md *ctx) {
    __u32 key = 0;
    __u64 *count = bpf_map_lookup_elem(&pkt_count, &key);
    if (count) {
        __sync_fetch_and_add(count, 1);
    }
    return XDP_PASS;
}

char __license[] SEC("license") = "GPL";
```

### 3. Generate Go Bindings

Create the `go:generate` directive:

```go
// xdp-counter/gen.go
package main

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go counter counter.c
```

Run code generation:

```bash
cd xdp-counter
go generate
```

This generates `counter_bpfel.go`, `counter_bpfeb.go`, and the corresponding `.o` object files.

### 4. Write the Go Userspace Loader

```go
// xdp-counter/main.go
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
)

func main() {
    // Load the compiled eBPF objects
    var objs counterObjects
    if err := loadCounterObjects(&objs, nil); err != nil {
        log.Fatalf("loading objects: %v", err)
    }
    defer objs.Close()

    // Get the network interface to attach to
    ifaceName := "lo" // Change to your interface (e.g., "eth0")
    iface, err := net.InterfaceByName(ifaceName)
    if err != nil {
        log.Fatalf("getting interface %s: %v", ifaceName, err)
    }

    // Attach the XDP program to the interface
    l, err := link.AttachXDP(link.XDPOptions{
        Program:   objs.CountPackets,
        Interface: iface.Index,
    })
    if err != nil {
        log.Fatalf("attaching XDP: %v", err)
    }
    defer l.Close()

    fmt.Printf("Counting packets on %s... Press Ctrl+C to stop.\n", ifaceName)

    // Graceful shutdown
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

    // Periodically read the packet count from the eBPF map
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            var count uint64
            key := uint32(0)
            if err := objs.PktCount.Lookup(key, &count); err != nil {
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
```

### 5. Build & Run

```bash
cd xdp-counter
go generate
go build -o counter .
sudo ./counter
```

> [!CAUTION]
> eBPF programs require **root privileges** (or `CAP_BPF` + `CAP_NET_ADMIN` capabilities) to load and attach.

---

## Key Concepts

### eBPF Program Types

| Type            | Hook Point             | Use Case                            |
| :-------------- | :--------------------- | :---------------------------------- |
| `XDP`           | Network driver (ingress)| Packet filtering, DDoS mitigation  |
| `TC`            | Traffic Control         | Packet mangling, policing          |
| `Tracepoint`    | Kernel tracepoints      | Syscall tracing, auditing          |
| `Kprobe/Kretprobe` | Kernel functions    | Dynamic tracing, debugging         |
| `Fentry/Fexit`  | Kernel functions (BTF)  | Faster kprobes with type safety    |
| `Cgroup`        | cgroup hooks            | Socket-level filtering & control   |
| `Perf Event`    | CPU perf counters       | Profiling, sampling                |

### eBPF Maps

Maps are key-value stores shared between the eBPF program (kernel) and the userspace program (Go):

| Map Type          | Description                                      |
| :---------------- | :----------------------------------------------- |
| `ARRAY`           | Fixed-size array indexed by integer key           |
| `HASH`            | Hash table with arbitrary keys                    |
| `PERF_EVENT_ARRAY`| Per-CPU ring buffer for streaming events          |
| `RING_BUF`        | Single-producer multi-consumer ring buffer        |
| `LRU_HASH`        | Hash with least-recently-used eviction            |
| `PERCPU_ARRAY`    | Per-CPU array (no locking needed)                 |

### Attach Points (Hooks)

```
Incoming Packet ──▶ XDP ──▶ TC ingress ──▶ Netfilter ──▶ Application
                                                              │
Application ──▶ TC egress ──▶ Outgoing Packet ◀──────────────┘
```

### bpf2go Code Generation

`bpf2go` is the recommended way to compile eBPF C programs and generate Go bindings:

```go
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target amd64 myprogram myprogram.c
```

This compiles `myprogram.c` → `myprogram_bpfel.o` and generates Go types/loaders.

---

## Examples

### Example 1: Counting Packets (XDP)

See [xdp-counter/](./xdp-counter/) — Attaches an XDP program to a network interface and counts every incoming packet.

### Example 2: Tracing Syscalls (Tracepoint)

See [tracepoint-openat/](./tracepoint-openat/) — Attaches to the `sys_enter_openat` tracepoint and logs which processes are opening files.

### Example 3: Monitoring TCP Connections (Kprobe)

See [kprobe-tcp/](./kprobe-tcp/) — Attaches a kprobe to `tcp_connect` and logs outgoing TCP connections with source/destination addresses.

---

## Debugging & Troubleshooting

```bash
# Check if eBPF is supported
bpftool feature

# List loaded eBPF programs
sudo bpftool prog list

# List eBPF maps
sudo bpftool map list

# Dump map contents
sudo bpftool map dump id <MAP_ID>

# View eBPF verifier output (on load failure)
# The Go loader will print verifier output automatically

# Check kernel version & config
uname -r
cat /boot/config-$(uname -r) | grep BPF
```

> [!TIP]
> Install `bpftool` for inspecting loaded eBPF programs and maps:
> ```bash
> sudo apt-get install -y linux-tools-$(uname -r)  # Ubuntu
> sudo dnf install -y bpftool                        # Fedora
> ```

---

## Resources

> [!IMPORTANT]
> **Documentation & References**
> - [ebpf-go Documentation](https://ebpf-go.dev/)
> - [ebpf-go Getting Started Guide](https://ebpf-go.dev/guides/getting-started/)
> - [cilium/ebpf GitHub](https://github.com/cilium/ebpf)
> - [cilium/ebpf Examples](https://github.com/cilium/ebpf/tree/main/examples)
> - [eBPF.io — What is eBPF?](https://ebpf.io/)
> - [BPF & XDP Reference Guide (Cilium)](https://docs.cilium.io/en/stable/bpf/)
> - [Liz Rice — Learning eBPF (O'Reilly)](https://www.oreilly.com/library/view/learning-ebpf/9781098135119/)
