# http-chunk-parse

按 HTTP/1.1 的分块传输编码解码正文：块长度十六进制、块扩展、可选 trailer，支持任意切分地喂入。只用标准库。

```
go test ./...
go vet ./...
go run ./cmd/chunked --sample=hex
go run ./cmd/chunked --sample=ext
go run ./cmd/chunked --sample=trailer
go run ./cmd/chunked --sample=badcrlf
go run ./cmd/chunked --sample=work
```

## 口径

- **块长**：每块以十六进制长度开头，行尾是 CRLF；长度与数据之间、数据与下一块之间都必须有 CRLF。
- **扩展**：长度后可带 `;` 开头的块扩展，解析长度时忽略扩展。
- **结束**：长度为 0 的块表示正文结束；其后到空行为止都是 trailer，不算正文。
- **任意切分**：可以按任意字节边界多次喂入，块头、块数据、结束段都可以跨次到达。

## 不变量

- 正文长度等于各块数据长度之和。
- 结束段之后的数据不再进入正文。
- `scanned` 不随喂入次数平方级放大：八百个块逐字节喂入的 `scanned` 不超过 6000。

## 输出契约

`Body` 返回正文；解析出错时 `Err` 非空，错误串为 `bad-size` 或 `bad-crlf`。
场景打印一行 JSON：出错打印 `{"error": ...}`，否则打印 `{"body": N}`。
`--sample=work` 打印 `{"chunks": N, "scanned": N}`。
