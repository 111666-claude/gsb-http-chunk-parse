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

type state int

const (
	stateHeader  state = iota // 读块长行
	stateData                 // 读块数据
	stateCRLF                 // 校验块尾 CRLF
	stateTrailer              // 读到结束段后的空行
	stateDone                 // 正文结束，后续数据全部丢弃
)

// Decoder 是分块解码器。
type Decoder struct {
	buf    []byte
	body   []byte
	err    error
	state  state
	remain int64
}

// NewDecoder 建一个解码器。
func NewDecoder() *Decoder {
	return &Decoder{}
}

// findCRLF 从 buf 的 pos 处找下一个 CRLF，找不到返回 -1。
func findCRLF(buf []byte, pos int, c *Counter) int {
	for i := pos; i+1 < len(buf); i++ {
		c.Scanned++
		if buf[i] == '\r' && buf[i+1] == '\n' {
			return i
		}
	}
	return -1
}

var (
	errBadSize = errors.New("bad-size")
	errBadCRLF = errors.New("bad-crlf")
)

// parseSize 解析块长行：十六进制长度，后可带 ';' 开头的块扩展。
func parseSize(line []byte) (int64, error) {
	if i := indexSemicolon(line); i >= 0 {
		line = line[:i]
	}
	if len(line) == 0 {
		return 0, errBadSize
	}
	size, err := strconv.ParseUint(string(line), 16, 63)
	if err != nil {
		return 0, errBadSize
	}
	return int64(size), nil
}

func indexSemicolon(line []byte) int {
	for i, b := range line {
		if b == ';' {
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
	pos := 0

loop:
	for d.err == nil {
		switch d.state {
		case stateHeader:
			idx := findCRLF(d.buf, pos, c)
			if idx < 0 {
				break loop
			}
			size, err := parseSize(d.buf[pos:idx])
			pos = idx + 2
			if err != nil {
				d.err = err
				break
			}
			if size == 0 {
				d.state = stateTrailer
			} else {
				d.remain = size
				d.state = stateData
			}
		case stateData:
			n := int64(len(d.buf) - pos)
			if n > d.remain {
				n = d.remain
			}
			d.body = append(d.body, d.buf[pos:pos+int(n)]...)
			pos += int(n)
			d.remain -= n
			if d.remain == 0 {
				d.state = stateCRLF
			} else {
				break loop
			}
		case stateCRLF:
			if len(d.buf)-pos < 2 {
				break loop
			}
			c.Scanned++
			if d.buf[pos] != '\r' || d.buf[pos+1] != '\n' {
				d.err = errBadCRLF
				break
			}
			pos += 2
			d.state = stateHeader
		case stateTrailer:
			idx := findCRLF(d.buf, pos, c)
			if idx < 0 {
				break loop
			}
			empty := idx == pos
			pos = idx + 2
			if empty {
				d.state = stateDone
			}
		case stateDone:
			pos = len(d.buf)
			break loop
		}
	}

	d.buf = d.buf[pos:]
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
