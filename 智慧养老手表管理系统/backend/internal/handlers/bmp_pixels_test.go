package handlers

import (
	"encoding/binary"
	"image"
	"testing"
)

func TestDecodeBMPUncompressedPixels(t *testing.T) {
	const width, height = 65, 65
	for _, channels := range []int{3, 4} {
		for _, topDown := range []bool{false, true} {
			stride := (width*channels + 3) &^ 3
			data := make([]byte, 54+stride*height)
			copy(data, "BM")
			binary.LittleEndian.PutUint32(data[10:], 54)
			binary.LittleEndian.PutUint32(data[14:], 40)
			binary.LittleEndian.PutUint32(data[18:], width)
			h := int32(height)
			if topDown {
				h = -h
			}
			binary.LittleEndian.PutUint32(data[22:], uint32(h))
			binary.LittleEndian.PutUint16(data[26:], 1)
			binary.LittleEndian.PutUint16(data[28:], uint16(8*channels))
			for y := 0; y < height; y++ {
				for x := 0; x < width; x++ {
					pixel := data[54+y*stride+x*channels:]
					pixel[0], pixel[1], pixel[2] = byte(x+y), byte(y), byte(x)
				}
			}
			img, w, hDecoded, ok := decodeBMP(data)
			if !ok || w != width || hDecoded != height {
				t.Fatalf("decode failed for channels=%d topDown=%t", channels, topDown)
			}
			rgba := img.(*image.RGBA)
			for y := 0; y < height; y++ {
				srcY := y
				if !topDown {
					srcY = height - 1 - y
				}
				for x := 0; x < width; x++ {
					got := rgba.RGBAAt(x, y)
					if got.R != byte(x) || got.G != byte(srcY) || got.B != byte(x+srcY) || got.A != 255 {
						t.Fatalf("pixel (%d,%d), channels=%d topDown=%t: %+v", x, y, channels, topDown, got)
					}
				}
			}
			if _, _, _, ok := decodeBMP(data[:len(data)-1]); ok {
				t.Fatal("accepted a truncated pixel buffer")
			}
		}
	}
}
