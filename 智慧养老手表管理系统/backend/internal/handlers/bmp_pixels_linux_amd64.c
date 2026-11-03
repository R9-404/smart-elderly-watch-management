//go:build cgo && !purego

#include <stddef.h>
#include <stdint.h>

// The caller has checked the sizes, row stride, and non-overlapping buffers.
void suining_unpack_bmp(uint8_t *restrict out, const uint8_t *restrict src,
                        int width, int height, int stride, int channels, int top_down) {
    for (int y = 0; y < height; ++y) {
        const int row = top_down ? y : height - 1 - y;
        const uint8_t *input = src + (size_t)row * stride;
        uint8_t *output = out + (size_t)y * width * 4;
        if (channels == 3) {
            for (int x = 0; x < width; ++x) {
                output[4*x] = input[3*x+2];
                output[4*x+1] = input[3*x+1];
                output[4*x+2] = input[3*x];
                output[4*x+3] = 255;
            }
        } else {
            for (int x = 0; x < width; ++x) {
                output[4*x] = input[4*x+2];
                output[4*x+1] = input[4*x+1];
                output[4*x+2] = input[4*x];
                output[4*x+3] = 255;
            }
        }
    }
}
