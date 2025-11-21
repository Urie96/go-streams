package streams

import (
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

type TestPacket struct {
	content string
	id      int
}

// 可以合并的包：相同ID的包可以合并
func mergeTestPackets(packetA, packetB TestPacket) (merged TestPacket, mergeable bool) {
	if packetA.id == packetB.id {
		return TestPacket{
			content: packetA.content + packetB.content,
			id:      packetA.id,
		}, true
	}
	return TestPacket{}, false
}

// 字符串包合并：总是可以合并
func mergeStrings(packetA, packetB string) (merged string, mergeable bool) {
	return packetA + packetB, true
}

// Batch merge function for ThrottleMerge2 - concatenates strings in batches
func batchMergeStrings(packets []string) []string {
	if len(packets) == 0 {
		return nil
	}
	// For this test, we'll concatenate all packets into one
	return []string{strings.Join(packets, "")}
}

// Batch merge function that creates multiple output packets from input
func batchMergeWithSize(packets []string) []string {
	var result []string
	// Split into chunks of 2 characters each
	for i := 0; i < len(packets); i += 2 {
		end := min(i+2, len(packets))
		result = append(result, strings.Join(packets[i:end], ""))
	}
	return result
}

func TestThrottleMerge2(t *testing.T) {
	t.Run("empty stream returns EOF immediately", func(t *testing.T) {
		src := Empty[string]()
		stream := ThrottleMerge2(src, batchMergeStrings, 10*time.Millisecond)
		expectStream(t, stream, []string{}, io.EOF)
	})

	t.Run("single packet without delay", func(t *testing.T) {
		packets := []string{"ab", "cd", "ef"}
		src := FromSlice(packets)
		stream := ThrottleMerge2(src, batchMergeStrings, 10*time.Millisecond)
		expectStream(t, stream, []string{"ab", "cdef"}, io.EOF)
	})

	t.Run("merge function returns multiple packets", func(t *testing.T) {
		packets := []string{"a", "b", "c", "d", "e", "f"}
		src := FromSlice(packets)
		stream := ThrottleMerge2(src, batchMergeWithSize, 20*time.Millisecond)

		expectStream(t, stream, []string{"a", "bc", "de", "f"}, io.EOF)
	})

	t.Run("slow input stream with throttle", func(t *testing.T) {
		packets := FromSlice([]string{"a", "b", "c"})
		src := FromFunc(func() (string, error) {
			time.Sleep(20 * time.Millisecond) // Slow producer
			return packets.Recv()
		})
		stream := ThrottleMerge2(src, batchMergeStrings, 10*time.Millisecond)

		// Should receive packets as they come, but merged after throttle
		expectStream(t, stream, []string{"a", "b", "c"}, io.EOF)
	})

	t.Run("error propagation after merge", func(t *testing.T) {
		streamErr := errors.New("stream error")
		src := Concat(FromSlice([]string{"hello", "world"}), FromErr[string](streamErr))
		stream := ThrottleMerge2(src, batchMergeStrings, 10*time.Millisecond)
		expectStream(t, stream, []string{"hello", "world"}, streamErr)
	})

	t.Run("immediate error without packets", func(t *testing.T) {
		streamErr := errors.New("immediate error")
		src := FromErr[string](streamErr)
		stream := ThrottleMerge2(src, batchMergeStrings, 10*time.Millisecond)
		expectStream(t, stream, []string{}, streamErr)
	})

	t.Run("zero throttle duration", func(t *testing.T) {
		packets := []string{"a", "b", "c"}
		src := FromSlice(packets)
		stream := ThrottleMerge2(src, batchMergeStrings, 0)
		expectStream(t, stream, []string{"a", "b", "c"}, io.EOF)
	})

	t.Run("complex batching scenario", func(t *testing.T) {
		packets := FromSlice([]string{"a", "b", "c", "d", "e", "f"})
		src := FromFunc(func() (string, error) {
			time.Sleep(50 * time.Millisecond) // Slow producer
			return packets.Recv()
		})
		stream := ThrottleMerge2(src, batchMergeStrings, 195*time.Millisecond)
		expectStream(t, stream, []string{"a", "bcde", "f"}, io.EOF)
	})
}

func TestThrottleMerge(t *testing.T) {
	t.Run("empty stream returns EOF immediately", func(t *testing.T) {
		src := Empty[TestPacket]()
		stream := ThrottleMerge(src, mergeTestPackets, 10*time.Millisecond)
		expectStream(t, stream, []TestPacket{}, io.EOF)
	})

	t.Run("single packet without delay returns immediately", func(t *testing.T) {
		packets := []TestPacket{{content: "hello", id: 1}}
		src := FromSlice(packets)
		stream := ThrottleMerge(src, mergeTestPackets, 10*time.Millisecond)

		expectStream(t, stream, []TestPacket{{content: "hello", id: 1}}, io.EOF)
	})

	t.Run("mergeable packets are merged", func(t *testing.T) {
		packets := []TestPacket{
			{content: "hello", id: 1},
			{content: "world", id: 1},
			{content: "foo", id: 1},
		}
		src := FromSlice(packets)
		stream := ThrottleMerge(src, mergeTestPackets, 10*time.Millisecond)

		expectStream(t, stream, []TestPacket{
			{content: "hello", id: 1},
			{content: "worldfoo", id: 1},
		}, io.EOF)
	})

	t.Run("non-mergeable packets are sent immediately", func(t *testing.T) {
		packets := []TestPacket{
			{content: "hello", id: 1},
			{content: "world", id: 2}, // Different ID, cannot merge
			{content: "foo", id: 3},   // Different ID, cannot merge
		}
		src := FromSlice(packets)
		stream := ThrottleMerge(src, mergeTestPackets, 10*time.Millisecond)

		expectStream(t, stream, []TestPacket{
			{content: "hello", id: 1},
			{content: "world", id: 2},
			{content: "foo", id: 3},
		}, io.EOF)
	})

	t.Run("throttle delay works", func(t *testing.T) {
		packets := []string{"a", "b", "c", "e", "f"}
		src := FromSlice(packets)
		stream := ThrottleMerge(src, mergeStrings, 50*time.Millisecond)
		start := time.Now()

		// First call returns immediately
		result, err := stream.Recv()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result != "a" {
			t.Errorf("Expected 'a', got '%v'", result)
		}
		firstCallDuration := time.Since(start)
		if firstCallDuration > 10*time.Millisecond {
			t.Errorf("First call should return immediately, took %v", firstCallDuration)
		}

		expectStream(t, stream, []string{"bcef"}, io.EOF)
	})

	t.Run("throttle timeout sends packet even if mergeable", func(t *testing.T) {
		packets := FromSlice([]string{"a", "b", "c", "d"})
		src := FromFunc(func() (string, error) {
			time.Sleep(20 * time.Millisecond) // Slow producer
			return packets.Recv()
		})
		stream := ThrottleMerge(src, mergeStrings, 10*time.Millisecond) // Short throttle

		expectStream(t, stream, []string{"a", "b", "c", "d"}, io.EOF)
	})

	t.Run("throttle merge", func(t *testing.T) {
		packets := FromSlice([]string{"a", "b", "c", "d", "e", "f"})
		src := FromFunc(func() (string, error) {
			time.Sleep(3 * time.Millisecond) // Slow producer
			return packets.Recv()
		})
		stream := ThrottleMerge(src, mergeStrings, 10*time.Millisecond) // Short throttle

		expectStream(t, stream, []string{"a", "bcd", "ef"}, io.EOF)
	})

	t.Run("stream error is propagated after buffer", func(t *testing.T) {
		streamErr := errors.New("stream error")
		packets := []TestPacket{{content: "hello", id: 1}}
		src := FromFunc(func() (TestPacket, error) {
			packet, err := FromSlice(packets).Recv()
			if err != nil {
				return packet, err
			}
			return packet, streamErr
		})
		stream := ThrottleMerge(src, mergeTestPackets, 10*time.Millisecond)

		expectStream(t, stream, []TestPacket{}, streamErr)
	})

	t.Run("complex merge scenario", func(t *testing.T) {
		packets := []TestPacket{
			{content: "a", id: 1}, // Start merge group 1
			{content: "b", id: 1}, // Can merge with previous
			{content: "x", id: 2}, // Cannot merge, send 'ab' first
			{content: "c", id: 1}, // Cannot merge with 'x', start new group
			{content: "d", id: 1}, // Can merge with previous
		}
		src := FromSlice(packets)
		stream := ThrottleMerge(src, mergeTestPackets, 10*time.Millisecond)

		expectStream(t, stream, []TestPacket{
			{content: "a", id: 1},
			{content: "b", id: 1},
			{content: "x", id: 2},
			{content: "cd", id: 1},
		}, io.EOF)
	})

	t.Run("string concatenation test", func(t *testing.T) {
		packets := []string{"hello", " ", "world", "!"}
		src := FromSlice(packets)
		stream := ThrottleMerge(src, mergeStrings, 10*time.Millisecond)

		expectStream(t, stream, []string{"hello", " world!"}, io.EOF)
	})

	t.Run("zero throttle duration", func(t *testing.T) {
		packets := []string{"a", "b", "c"}
		src := FromSlice(packets)
		stream := ThrottleMerge(src, mergeStrings, 0)

		expectStream(t, stream, []string{"a", "b", "c"}, io.EOF)
	})
}
