//go:build ignore

#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

// Define the format of the tracepoint struct (simplified)
struct trace_event_raw_sys_enter_openat {
    __u64 unused;
    int dfd;
    const char *filename;
    int flags;
    __u16 mode;
};

// Map to count how many times each process (PID) calls openat
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, __u32);
    __type(value, __u64);
    __uint(max_entries, 10240);
} openat_counts SEC(".maps");

SEC("tracepoint/syscalls/sys_enter_openat")
int trace_openat(struct trace_event_raw_sys_enter_openat *ctx) {
    __u32 pid = bpf_get_current_pid_tgid() >> 32;
    __u64 *count = bpf_map_lookup_elem(&openat_counts, &pid);
    
    if (count) {
        __sync_fetch_and_add(count, 1);
    } else {
        __u64 init_val = 1;
        bpf_map_update_elem(&openat_counts, &pid, &init_val, BPF_ANY);
    }
    return 0;
}

char __license[] SEC("license") = "GPL";
