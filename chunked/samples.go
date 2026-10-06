package chunked

import "fmt"

// Scenario 是一个固定场景。
type Scenario struct {
	Name string
	Run  func() map[string]any
}

// Encode 把若干块拼成分块编码流。
func Encode(chunks [][]byte) []byte {
	var out []byte
	for _, ch := range chunks {
		out = append(out, []byte(fmt.Sprintf("%x\r\n", len(ch)))...)
		out = append(out, ch...)
		out = append(out, '\r', '\n')
	}
	return append(out, []byte("0\r\n\r\n")...)
}

func one(data []byte) map[string]any {
	d := NewDecoder()
	err := d.Feed(data, &Counter{})
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	if d.err != nil {
		return map[string]any{"error": d.err.Error()}
	}
	return map[string]any{"body": len(d.Body())}
}

// Samples 返回全部场景。
func Samples() []Scenario {
	return []Scenario{
		{Name: "hex", Run: func() map[string]any {
			return one(Encode([][]byte{[]byte("abcdefghijklmnopqrstuvwxyz")}))
		}},
		{Name: "ext", Run: func() map[string]any {
			return one([]byte("5;a=b\r\nhello\r\n0\r\n\r\n"))
		}},
		{Name: "trailer", Run: func() map[string]any {
			return one([]byte("5\r\nhello\r\n0\r\nX: 1\r\n\r\n"))
		}},
		{Name: "badcrlf", Run: func() map[string]any {
			return one([]byte("5\r\nhelloZZ0\r\n\r\n"))
		}},
	}
}

// Work 跑规模线场景：逐字节喂入。
func Work(n int) map[string]any {
	chunks := make([][]byte, n)
	for i := range chunks {
		chunks[i] = []byte("x")
	}
	data := Encode(chunks)
	d := NewDecoder()
	c := &Counter{}
	for _, b := range data {
		if err := d.Feed([]byte{b}, c); err != nil {
			break
		}
		if d.err != nil {
			break
		}
	}
	return map[string]any{"chunks": n, "scanned": c.Scanned}
}
