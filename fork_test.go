package streams

import (
	"fmt"
	"io"
	"sync"
	"testing"
)

func TestTeeReader(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		a, b := TeeReader(FromSlice([]int{1, 2, 3}))

		// 交替读取
		v1, err := a.Recv()
		if err != nil || v1 != 1 {
			t.Fatalf("a.Recv() = %v, %v; want 1, nil", v1, err)
		}
		v2, err := b.Recv()
		if err != nil || v2 != 1 {
			t.Fatalf("b.Recv() = %v, %v; want 1, nil", v2, err)
		}

		v3, err := a.Recv()
		if err != nil || v3 != 2 {
			t.Fatalf("a.Recv() = %v, %v; want 2, nil", v3, err)
		}
		v4, err := b.Recv()
		if err != nil || v4 != 2 {
			t.Fatalf("b.Recv() = %v, %v; want 2, nil", v4, err)
		}

		v5, err := a.Recv()
		if err != nil || v5 != 3 {
			t.Fatalf("a.Recv() = %v, %v; want 3, nil", v5, err)
		}
		v6, err := b.Recv()
		if err != nil || v6 != 3 {
			t.Fatalf("b.Recv() = %v, %v; want 3, nil", v6, err)
		}

		// 读到 EOF
		_, err = a.Recv()
		if err != io.EOF {
			t.Fatalf("a.Recv() err = %v; want EOF", err)
		}
		_, err = b.Recv()
		if err != io.EOF {
			t.Fatalf("b.Recv() err = %v; want EOF", err)
		}
	})

	t.Run("concurrent", func(t *testing.T) {
		a, b := TeeReader(FromSlice([]int{1, 2, 3, 4}))
		c, d := TeeReader(b)

		errCh := make(chan error, 10)
		var wg sync.WaitGroup
		wg.Add(3)

		validate := func(s Stream[int]) {
			defer wg.Done()
			for want := 1; want <= 4; want++ {
				v, err := s.Recv()
				if err != nil {
					errCh <- err
					return
				}
				if v != want {
					errCh <- fmt.Errorf("got %d, want %d", v, want)
					return
				}
			}
			if _, err := s.Recv(); err != io.EOF {
				errCh <- fmt.Errorf("expected EOF, got %v", err)
			}
		}

		go validate(a)
		go validate(c)
		go validate(d)

		wg.Wait()
		close(errCh)

		for err := range errCh {
			if err != nil {
				t.Fatal(err)
			}
		}
	})
}
