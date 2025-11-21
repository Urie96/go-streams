package streams

import (
	"testing"
)

func TestWithLog(t *testing.T) {
	logger := func(info string) { t.Log(info) }
	stream := WithLog(FromSlice([]string{"a", "b", "c"}), "test", logger)

	res, err := CollectString(stream)
	if err != nil {
		t.Errorf("err is %v", err)
	}
	if res != "abc" {
		t.Errorf("res is %s", res)
	}
}
