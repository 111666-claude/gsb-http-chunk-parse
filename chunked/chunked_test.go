package chunked

import "testing"

func TestSingleChunk(t *testing.T) {
	d := NewDecoder()
	if err := d.Feed([]byte("5\r\nhello\r\n0\r\n\r\n"), &Counter{}); err != nil {
		t.Fatalf("解码失败：%v", err)
	}
	if string(d.Body()) != "hello" {
		t.Fatalf("正文应是 hello：%q", d.Body())
	}
}

func TestEmptyBody(t *testing.T) {
	d := NewDecoder()
	d.Feed([]byte("0\r\n\r\n"), &Counter{})
	if len(d.Body()) != 0 {
		t.Fatalf("空流不该有正文：%q", d.Body())
	}
}

func TestMultipleChunks(t *testing.T) {
	d := NewDecoder()
	d.Feed(Encode([][]byte{[]byte("abc"), []byte("def")}), &Counter{})
	if string(d.Body()) != "abcdef" {
		t.Fatalf("正文应是 abcdef：%q", d.Body())
	}
}

func TestNoErrorOnValid(t *testing.T) {
	d := NewDecoder()
	if err := d.Feed(Encode([][]byte{[]byte("abc")}), &Counter{}); err != nil {
		t.Fatalf("不该报错：%v", err)
	}
}

func TestCounterCounts(t *testing.T) {
	c := &Counter{}
	NewDecoder().Feed(Encode([][]byte{[]byte("abc")}), c)
	if c.Scanned == 0 {
		t.Fatal("应统计比较次数")
	}
}

func TestBodyLength(t *testing.T) {
	d := NewDecoder()
	d.Feed(Encode([][]byte{[]byte("abcde")}), &Counter{})
	if len(d.Body()) != 5 {
		t.Fatalf("正文长度应是 5：%d", len(d.Body()))
	}
}
