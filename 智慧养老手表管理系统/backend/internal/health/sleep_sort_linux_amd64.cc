//go:build cgo && !purego

#include <algorithm>
#include <cstdint>

extern "C" {
struct suining_sleep_key {
    int64_t seconds;
    uint32_t nanos;
    uint32_t index;
};

void suining_sort_sleep(suining_sleep_key *rows, uint32_t count) {
    std::sort(rows, rows + count, [](const suining_sleep_key &a, const suining_sleep_key &b) {
        return a.seconds < b.seconds || (a.seconds == b.seconds && a.nanos < b.nanos);
    });
}
}
