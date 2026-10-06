// Package chunked 解码分块传输编码。
package chunked

import (
	"errors"
	"strconv"
)

// Counter 记录比较次数。
type Counter struct {
	Scanned int
}

// Decoder 是分块解码器。
type Decoder struct {
	buf  []byte
	pos  int
	body []byte
	err  error
}

// NewDecoder 建一个解码器。
func NewDecoder() *Decoder {
	return &Decoder{}
}

func indexCRLF(buf []byte, pos int, c *Counter) int {
	for i := pos; i+1 < len(buf); i++ {
		c.Scanned++
		if buf[i] == '\r' && buf[i+1] == '\n' {
			return i
		}
	}
	return -1
}

// Feed 喂入一段字节。
func (d *Decoder) Feed(p []byte, c *Counter) error {
	d.buf = append(d.buf, p...)
	body := []byte{}
	pos := 0
	for pos < len(d.buf) {
		idx := indexCRLF(d.buf, pos, c)
		if idx < 0 {
			break
		}
		line := string(d.buf[pos:idx])
		pos = idx + 2
		size, err := strconv.Atoi(line)
		if err != nil {
			d.err = errors.New("bad-size")
			return d.err
		}
		if size == 0 {
			for {
				idx2 := indexCRLF(d.buf, pos, c)
				if idx2 < 0 {
					break
				}
				line2 := string(d.buf[pos:idx2])
				pos = idx2 + 2
				if line2 == "" {
					break
				}
				body = append(body, line2...)
			}
			break
		}
		end := pos + size
		if end > len(d.buf) {
			break
		}
		body = append(body, d.buf[pos:end]...)
		pos = end + 2
	}
	d.body = body
	return d.err
}

// Body 返回已解码正文。
func (d *Decoder) Body() []byte {
	return d.body
}

// Err 返回解析错误。
func (d *Decoder) Err() error {
	return d.err
}
