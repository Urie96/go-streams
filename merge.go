package streams

import (
	"time"
)

// ThrottleMerge 将流里面的多个包进行聚合，然后再发送给下游。用于sse攒包推送
// 参数merge：merge函数用于判断两个包是否可以合并，如果可以合并，则返回合并后的包和true，否则返回false，该包会立即发给下游
// 参数throttleDuration：如果当前的包都可以合并，那么等待这段时间之后再发送
// 首包会立即发送，后面的包会根据throttleDuration进行聚合
func ThrottleMerge[T any](s Stream[T], merge func(packetA, packetB T) (merged T, mergeable bool), throttleDuration time.Duration) Stream[T] {
	var zero T
	var buf *T
	var bufErr error
	lastSendAt := time.Time{}

	sendBuf := func() T {
		lastBuf := *buf
		buf = nil
		lastSendAt = time.Now()
		return lastBuf
	}

	return FromFunc(func() (T, error) {
		if bufErr != nil {
			return zero, bufErr
		}
		if buf == nil {
			packet, err := s.Recv()
			if err != nil {
				return packet, err
			}
			buf = &packet
		}
		for {
			if time.Since(lastSendAt) > throttleDuration {
				return sendBuf(), nil
			} else {
				packet, err := s.Recv()
				if err != nil {
					bufErr = err
					return sendBuf(), nil
				}
				merged, mergeable := merge(*buf, packet)
				if !mergeable {
					toSend := sendBuf()
					buf = &packet // 下次再merge
					return toSend, nil
				} else {
					buf = &merged
				}
			}
		}
	})
}

// ThrottleMerge2 每隔指定时间，将流里面的多个包进行聚合成新的包，然后再发送给下游。用于sse攒包推送
// merge函数用于将m个包合并成n个包发给下游
// 首包会立即发送，后面的包会根据throttleDuration进行聚合
func ThrottleMerge2[T any](s Stream[T], merge func(packets []T) []T, throttleDuration time.Duration) Stream[T] {
	var zero T
	var origBuf, sendBuf []T
	var sendErr error
	lastMergeAt := time.Time{}

	mergeOrigBuf := func() {
		lastMergeAt = time.Now()
		if len(origBuf) > 0 {
			sendBuf = append(sendBuf, merge(origBuf)...)
			origBuf = origBuf[:0]
		}
	}

	receiveOrigBuf := func() {
		packet, err := s.Recv()
		if err != nil {
			mergeOrigBuf()
			sendErr = err
		} else {
			origBuf = append(origBuf, packet)
		}
	}

	return FromFunc(func() (T, error) {
		for {
			if len(sendBuf) > 0 { // 如果有缓存，直接发送
				send := sendBuf[0]
				sendBuf = sendBuf[1:]
				return send, nil
			}
			if sendErr != nil { // 缓存已空，检查缓存的错误，直接发送
				return zero, sendErr
			}

			if time.Since(lastMergeAt) > throttleDuration {
				if len(origBuf) == 0 {
					receiveOrigBuf()
				}
				mergeOrigBuf()
			} else {
				receiveOrigBuf()
			}
		}
	})
}
