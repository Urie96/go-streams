package streams

import (
	"fmt"
	"runtime"
	"testing"
	"time"
	"unsafe"
)

func newStream() chan [1024]int {
	ch := make(chan [1024]int, 100)
	go func() {
		for range 100000 {
			ch <- [1024]int{}
		}
		close(ch)
	}()
	return ch
}

func testBufferGC() {
	src := FromChan(newStream())
	src = WithBuffer(src)
	printMemStats()
	chunk, _ := src.Recv()
	fmt.Println("首包大小", unsafe.Sizeof(chunk))
}

func printMemStats() {
	runtime.GC()
	time.Sleep(time.Second)
	runtime.GC()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	fmt.Printf("堆内存: %d kB\n", memStats.HeapAlloc/8/1024)
}

func TestWithBuffer(t *testing.T) {
	testBufferGC()
	printMemStats()
}
