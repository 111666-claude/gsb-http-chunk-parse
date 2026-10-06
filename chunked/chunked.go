// Package chunked 解码分块传输编码。
package chunked

import (
	"bytes"
	"errors"
	"strconv"
)

// Counter 记录比较次数。
type Counter struct {
	Scanned int
}

// Decoder 是分块解码器。
type Decoder struct {
	buf   []byte
	size  int
	body  []byte
	err   error
	state int
}

const (
	stateSize = iota
	stateData
	stateTrailer
	stateDone
)

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
	if d.err != nil {
		return d.err
	}
	d.buf = append(d.buf, p...)
	for d.state != stateDone {
		switch d.state {
		case stateSize:
			idx := indexCRLF(d.buf, 0, c)
			if idx < 0 {
				return nil
			}
			line := d.buf[:idx]
			d.buf = d.buf[idx+2:]
			if i := bytes.IndexByte(line, ';'); i >= 0 {
				line = line[:i]
			}
			n, err := strconv.ParseUint(string(line), 16, 63)
			if err != nil {
				d.err = errors.New("bad-size")
				return d.err
			}
			d.size = int(n)
			if n == 0 {
				d.state = stateTrailer
			} else {
				d.state = stateData
			}
		case stateData:
			if len(d.buf) < d.size+2 {
				return nil
			}
			if d.buf[d.size] != '\r' || d.buf[d.size+1] != '\n' {
				d.err = errors.New("bad-crlf")
				return d.err
			}
			d.body = append(d.body, d.buf[:d.size]...)
			d.buf = d.buf[d.size+2:]
			d.state = stateSize
		case stateTrailer:
			idx := indexCRLF(d.buf, 0, c)
			if idx < 0 {
				return nil
			}
			empty := idx == 0
			d.buf = d.buf[idx+2:]
			if empty {
				d.state = stateDone
			}
		}
	}
	d.buf = nil
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
