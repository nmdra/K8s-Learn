//go:build ignore

#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

// A ring buffer map for sending events to userspace
struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1 << 24);
} events SEC(".maps");

// Event structure sent to userspace
struct event_t {
    __u32 pid;
    char comm[16];
};

SEC("kprobe/tcp_v4_connect")
int kprobe_tcp_v4_connect(void *ctx) {
    struct event_t *event;

    // Reserve space in the ring buffer
    event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
    if (!event) {
        return 0; // Buffer full
    }

    // Populate event data
    event->pid = bpf_get_current_pid_tgid() >> 32;
    bpf_get_current_comm(&event->comm, sizeof(event->comm));

    // Submit event to userspace
    bpf_ringbuf_submit(event, 0);

    return 0;
}

char __license[] SEC("license") = "GPL";
