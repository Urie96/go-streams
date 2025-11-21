package streams

import (
	"io"
	"strings"
	"testing"
)

func TestLabelStream_Recv(t *testing.T) {
	newTestStream := func(whole string) Stream[string] {
		return FromSlice(strings.Split(whole, ""))
	}

	t.Run("正常流程", func(t *testing.T) {
		stream := newTestStream(`我是第一段测试我是第一段测试我是第一段测试我是第一段测试我是第一段测试
### 【结论判断】
本结论整合1份2019年7月5日的右乳癌术后复查报告，结合患者病情及病史，综合判断为右乳癌术后、右上肺尖段类结节。
- **疾病判断**：右乳癌术后、右上肺尖段类结节
- **判断依据**：
    - 右乳癌术后复查，右乳缺失，局部未见明确复发征象。

### 【补充上传】
上传近期复查报告或历史报告获得更完整解读`)
		ls := NewLabelStream(stream, []SLabel{
			{
				Name:       "first_line",
				StartToken: "",
				EndToken:   "\n",
			},
			{
				Name:       "conclusion",
				StartToken: "### 【结论判断】\n",
				EndToken:   "\n",
			},
			{
				Name:       "title",
				StartToken: "- **疾病判断**",
				EndToken:   "\n",
			},
			{
				Name:       "proof",
				StartToken: "- **判断依据**：\n",
				EndToken:   "\n\n",
			},
			{
				Name:       "upload_more",
				StartToken: "### 【补充上传】\n",
				EndToken:   "",
			},
		})
		Consume(ls, func(frame LabeledChunk) error {
			t.Log(frame.Label, "<"+frame.Chunk+">")
			return nil
		})
	})
}

func TestLabelStream_Split(t *testing.T) {
	t.Parallel()
	src := FromSlice(strings.Split("hello <A>chunk1</A> world <B>chunk2</B> tail", ""))
	labels := []SLabel{
		{Name: "A", StartToken: "<A>", EndToken: "</A>"},
		{Name: "B", StartToken: "<B>", EndToken: "</B>"},
	}
	ls := NewLabelStream(src, labels)
	demux := ls.Demux()

	t.Run("label A", func(t *testing.T) {
		t.Parallel()
		expectStringStream(t, demux["A"], "chunk1", io.EOF)
	})

	t.Run("label B", func(t *testing.T) {
		t.Parallel()
		expectStringStream(t, demux["B"], "chunk2", io.EOF)
	})

	t.Run("default label", func(t *testing.T) {
		t.Parallel()
		expectStringStream(t, demux[""], "hello </A> world </B> tail", io.EOF)
	})
}

func TestLabelStream_Split_NoEndToken(t *testing.T) {
	src := FromSlice([]string{"start", "<X>rest", "of", "stream"})
	labels := []SLabel{
		{Name: "X", StartToken: "<X>", EndToken: ""},
	}
	ls := NewLabelStream(src, labels)
	demux := ls.Demux()
	expectStringStream(t, demux["X"], "restofstream", io.EOF)
}
