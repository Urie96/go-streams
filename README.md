# Streams

为流式数据读取提供了一系列**惰性求值**的工具函数，不涉及任何 goroutine 和 channel，支持对流部分消费。

> [!TIP]
> 惰性求值：只有在消费流时才会执行中间处理。
>
> 使用chan来承载流式数据读取的痛点：
>
> 1. 需要在生产方起一个goroutine循环写入chan，所以要求消费方必须消费完chan里面的所有数据，否则可能使得生成方goroutine泄露；
> 2. 如果在消费方的中间态过程处理chan的数据，在协程中全量求值，只能起一个goroutine循环处理并返回一个新chan，同样，也要求下游必须消费所有chan;
> 3. 多协程+channel可能导致频繁的协程调度开销，小流量难以捕捉的并发安全问题及死锁风险等;

## 快速开始

```go
import "github.com/urie96/go-streams"

type Stream[T any] interface { // StreamReader
	Recv() (T, error)
}

// =====================================生产 ====================================
stream, err := ai_model_proxy.ChatBizStreaming(ctx, req) // rpc流式响应
if err != nil {
	return streams.FromErr[*ChatResp](err)
}
sliceStream := streams.FromSlice([]string{"hello", "world"}) // 从切片创建流
chStream := streams.FromChan(make(chan string))              // 从channel创建流

// =====================================中间处理 ====================================
s := streams.Map(stream, func(c *ai_model_proxy.ChatBizStreamingResponse) string {
	return c.GetMessage()
}) // 将原流中每个数据转换为字符串并返回新流
s = streams.RemoveLabels(s, []string{"<|diagnosis|>", "<|/diagnosis|>"}) // 返回一个新流，移除流中的指定标签
labelStreams := streams.NewLabelStream(s, []streams.SLabel{
	{Name: "answer", StartToken: "", EndToken: "<|inquiry|>"},
	{Name: "question", StartToken: "<|inquiry|>", EndToken: "<|/inquiry|>"},
}).Demux() // 将一个流拆分为多个流，每个流包含一个label
answerStream := labelStreams["answer"]                                    // inquiry特殊标记前的流
questionStream := labelStreams["question"]                                // inquiry特殊标记之间的流
defaultStream := labelStreams[""]                                         // 未被label标记的流，在此例中即inquiry结束标记之后的流
concatStream := streams.Concat(answerStream, questionStream)              // 合并多个流，前一个流结束之后，继续输出第二个流
firstNonEmptyStream := streams.FirstNonEmpty(concatStream, defaultStream) // 从多个流中获取第一个非空的流
s1, s2 := streams.TeeReader(firstNonEmptyStream)                          // 将流按原样复制为两个流
s1 = streams.WithLog(s1, "test-key", func(info string) { t.Log(info) })   // 注入log函数，在未来，流开始和结束时分别打印一次日志
s2 = streams.ToSafe(s2)                                                   // 将未知流转换为并发安全的流，如果流在下游被并发消费时有用

// =====================================消费 ====================================
s2 = streams.NewStringReader(s2)
answer := s2.ReadUntil([]string{"<|inquiry|>"})       // 阻塞收集流中的数据，直到遇到<|inquiry|>
question := streams.CollectString(s2)                 // 阻塞收集流中的所有数据
streams.Consume(s1, func(s string) error { send(s) }) // 阻塞，对流中的每个数据调用一次send函数
for chunk := range streams.Iter(s1) {                 // 迭代器，在agent服务scm go版本升级到1.23之前不可用
	send(chunk)
}
for chunk := range streams.ToChan(s1) { // 流转为channel，内部起一个go协程循环读取并写入chan，要求下游需要将chan消费完毕，否则导致协程泄露
    send(chunk)
}
```

### 中间处理

#### `NewLabelStream`

`NewLabelStream` 用于给一个string流的每个数据打标签，调用Demux可以根据标签拆分为多条流。

```go
src := FromSlice(strings.Split("hello <A>chunk1</A> world <B>chunk2</B> tail", ""))
labels := []SLabel{
	{Name: "A", StartToken: "<A>", EndToken: "</A>"},
	{Name: "B", StartToken: "<B>", EndToken: "</B>"},
}
ls := NewLabelStream(src, labels)
demux := ls.Demux()
expectStringStream(t, demux["A"], "chunk1", io.EOF)
expectStringStream(t, demux["B"], "chunk2", io.EOF)
expectStringStream(t, demux[""], "hello </A> world </B> tail", io.EOF)
```

#### `NewSpecialTokenParserStream`

`NewSpecialTokenParserStream` 用于给一个string流的每个数据打标签，调用Demux可以根据标签拆分为多条流。

```go
src := FromSlice([]string{"sta", "rt<|inq", "uiry|>h", "ello", "world", "<|/inq", "uiry|>", "end", "sdf<|diag", "nosis|>", "stop"})
demux := NewSpecialTokenParserStream(src, []string{"<|inquiry|>", "<|/inquiry|>", "<|diagnosis|>"}).Demux()
expectStringStream(t, demux[""], "start", io.EOF)
expectStringStream(t, demux["<|inquiry|>"], "helloworld", io.EOF)
expectStringStream(t, demux["<|/inquiry|>"], "endsdf", io.EOF)
expectStringStream(t, demux["<|diagnosis|>"], "stop", io.EOF)
```

#### `SubstituteStream`

`SubstituteStream` 用于当流消费到特定数据时，替换为一个新流供下游继续消费。

```go
substituteStream := SubstituteStream(FromSlice([]string{"start", "<|diagnosis|>", "end"}), func(item string) Stream[string] {
	if item == "<|diagnosis|>" {
		return FromSlice([]string{"我是", "咨询小结"})
	}
	return nil
})
expectStream(t, substituteStream, []string{"start", "我是", "咨询小结"}, io.EOF)
```
