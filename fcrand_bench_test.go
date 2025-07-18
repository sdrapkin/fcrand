package fcrand

import (
	gorand "crypto/rand"
	"fmt"
	"strconv"
	"testing"
)

// global
var _sizes []int = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 12, 16, 17, 25, 31, 32, 33, 48, 56, 57, 64, 65, 128, 256, 512, 513, 1024, 2048, 4096}
var _goroutineCounts []int = []int{64} // []int{2, 4, 8, 16, 32, 64}

func Benchmark_fcrand_Serial(b *testing.B) {
	b.ReportAllocs()
	for _, size := range _sizes {
		benchName := "Size_" + strconv.Itoa(size)
		buf := make([]byte, size)
		if size > 0 {
			for range 1024 {
				Read(buf)
			}
		}
		b.Run(benchName, func(b *testing.B) {
			b.SetBytes(int64(size))
			b.ResetTimer()
			for b.Loop() {
				_, err := Read(buf)
				if err != nil {
					b.Fatalf("Read failed: %v", err)
				}
			}
		})
	}
}

func Benchmark_gorand_Serial(b *testing.B) {
	b.ReportAllocs()
	for _, size := range _sizes {
		benchName := "Size_" + strconv.Itoa(size)
		buf := make([]byte, size)
		if size > 0 {
			for range 1024 {
				gorand.Read(buf)
			}
		}
		b.Run(benchName, func(b *testing.B) {
			b.SetBytes(int64(size))
			b.ResetTimer()
			for b.Loop() {
				_, err := gorand.Read(buf)
				if err != nil {
					b.Fatalf("Read failed: %v", err)
				}
			}
		})
	}
}

func Benchmark_fcrand_Concur(b *testing.B) {
	b.ReportAllocs()
	for _, size := range _sizes {
		for _, count := range _goroutineCounts {
			benchName := fmt.Sprintf("Size_%d_G%d", size, count)
			b.Run(benchName, func(b *testing.B) {
				if size > 0 {
					buf := make([]byte, size)
					for range 1024 {
						Read(buf)
					}
				}
				b.SetBytes(int64(size))
				b.SetParallelism(count)
				b.ResetTimer()
				b.RunParallel(func(pb *testing.PB) {
					buf := make([]byte, size)
					for pb.Next() {
						_, err := Read(buf)
						if err != nil {
							b.Fatalf("Read failed: %v", err)
						}
					}
				})
			})
		}
	}
}

func Benchmark_gorand_Concur(b *testing.B) {
	b.ReportAllocs()
	for _, size := range _sizes {
		for _, count := range _goroutineCounts {
			benchName := fmt.Sprintf("Size_%d_G%d", size, count)
			b.Run(benchName, func(b *testing.B) {
				if size > 0 {
					buf := make([]byte, size)
					for range 1024 {
						gorand.Read(buf)
					}
				}
				b.SetBytes(int64(size))
				b.SetParallelism(count)
				b.ResetTimer()
				b.RunParallel(func(pb *testing.PB) {
					buf := make([]byte, size)
					for pb.Next() {
						_, err := gorand.Read(buf)
						if err != nil {
							b.Fatalf("Read failed: %v", err)
						}
					}
				})
			})
		}
	}
}
