// Package splitter 提供文本分块工具，支持按字符数、递归分隔符、Token 数三种策略.
package splitter

import (
	"strings"
	"unicode"
)

// Splitter 文本分块器接口.
type Splitter interface {
	Split(text string) []Chunk
}

// Chunk 文本块.
type Chunk struct {
	Text     string         `json:"text"`
	Offset   int            `json:"offset"`            // 原文起始位置（字节偏移）
	Index    int            `json:"index"`             // 块序号（从 0 开始）
	Metadata map[string]any `json:"metadata,omitzero"` // 附加元数据
}

// Option 分块器配置项.
type Option func(*options)

type options struct {
	chunkSize    int
	chunkOverlap int
	separators   []string
}

func defaultOptions() options {
	return options{
		chunkSize:    1000,
		chunkOverlap: 200,
		separators:   []string{"\n\n", "\n", "。", ".", " ", ""},
	}
}

// WithChunkSize 设置每块的最大字符数（默认 1000）.
func WithChunkSize(n int) Option {
	return func(o *options) { o.chunkSize = n }
}

// WithChunkOverlap 设置相邻块之间的重叠字符数（默认 200）.
func WithChunkOverlap(n int) Option {
	return func(o *options) { o.chunkOverlap = n }
}

// WithSeparators 设置递归分块时使用的分隔符列表（优先级从高到低）.
func WithSeparators(seps []string) Option {
	return func(o *options) { o.separators = seps }
}

// applyOptions 将 Option 列表应用到默认配置上.
func applyOptions(opts []Option) options {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

// ──────────────────────────────────────────
// CharacterSplitter
// ──────────────────────────────────────────

type characterSplitter struct {
	opts options
}

// NewCharacterSplitter 按字符数分块，相邻块保留 chunkOverlap 字符的重叠内容.
func NewCharacterSplitter(opts ...Option) Splitter {
	return &characterSplitter{opts: applyOptions(opts)}
}

func (s *characterSplitter) Split(text string) []Chunk {
	if text == "" {
		return nil
	}
	runes := []rune(text)
	size := s.opts.chunkSize
	overlap := s.opts.chunkOverlap
	if overlap >= size {
		overlap = size / 2
	}

	var chunks []Chunk
	start := 0
	for start < len(runes) {
		end := start + size
		if end > len(runes) {
			end = len(runes)
		}
		chunk := string(runes[start:end])
		// 计算字节偏移
		byteOffset := len(string(runes[:start]))
		chunks = append(chunks, Chunk{
			Text:   chunk,
			Offset: byteOffset,
			Index:  len(chunks),
		})
		if end == len(runes) {
			break
		}
		start += size - overlap
	}
	return chunks
}

// ──────────────────────────────────────────
// RecursiveSplitter
// ──────────────────────────────────────────

type recursiveSplitter struct {
	opts options
}

// NewRecursiveSplitter 递归分块（段落→句子→字符）.
// 默认分隔符顺序: ["\n\n", "\n", "。", ".", " ", ""]
func NewRecursiveSplitter(opts ...Option) Splitter {
	return &recursiveSplitter{opts: applyOptions(opts)}
}

func (s *recursiveSplitter) Split(text string) []Chunk {
	if text == "" {
		return nil
	}
	// 递归收集所有文本片段及其原始偏移
	pieces := recursiveSplit(text, 0, s.opts.separators, s.opts.chunkSize)
	// 合并过短的片段，同时维护 offset
	merged := mergeChunks(pieces, s.opts.chunkSize, s.opts.chunkOverlap)
	// 赋予连续的 Index
	for i := range merged {
		merged[i].Index = i
	}
	return merged
}

// piece 是递归分割过程中的中间表示，保留字节偏移.
type piece struct {
	text   string
	offset int // 字节偏移（相对于原始文本）
}

// recursiveSplit 对 text（其起始字节偏移为 baseOffset）按分隔符列表递归分割.
func recursiveSplit(text string, baseOffset int, seps []string, maxSize int) []piece {
	if len(seps) == 0 {
		// 兜底：按字符逐一返回
		var out []piece
		for i, r := range text {
			out = append(out, piece{text: string(r), offset: baseOffset + i})
		}
		return out
	}

	sep := seps[0]
	restSeps := seps[1:]

	var parts []string
	if sep == "" {
		// 按单个字符拆分
		for _, r := range text {
			parts = append(parts, string(r))
		}
	} else {
		parts = strings.Split(text, sep)
	}

	var pieces []piece
	curOffset := baseOffset
	for i, part := range parts {
		if part == "" {
			// 空串跳过，但需要推进偏移（分隔符本身）
			if i < len(parts)-1 {
				curOffset += len(sep)
			}
			continue
		}
		if len([]rune(part)) <= maxSize {
			pieces = append(pieces, piece{text: part, offset: curOffset})
		} else {
			// 该片段仍然过大，用下一个分隔符继续拆分
			sub := recursiveSplit(part, curOffset, restSeps, maxSize)
			pieces = append(pieces, sub...)
		}
		curOffset += len(part)
		if i < len(parts)-1 {
			curOffset += len(sep) // 加上分隔符本身的长度
		}
	}
	return pieces
}

// mergeChunks 将 piece 列表按 chunkSize / chunkOverlap 合并为最终 Chunk 列表.
func mergeChunks(pieces []piece, maxSize, overlap int) []Chunk {
	if len(pieces) == 0 {
		return nil
	}
	if overlap >= maxSize {
		overlap = maxSize / 2
	}

	var chunks []Chunk
	var buf []piece // 当前窗口内的 piece
	bufSize := 0    // buf 中所有文本的 rune 数

	flush := func() {
		if len(buf) == 0 {
			return
		}
		var sb strings.Builder
		for _, p := range buf {
			sb.WriteString(p.text)
		}
		chunks = append(chunks, Chunk{
			Text:   sb.String(),
			Offset: buf[0].offset,
		})
	}

	for _, p := range pieces {
		pSize := len([]rune(p.text))
		if bufSize+pSize > maxSize && len(buf) > 0 {
			flush()
			// 保留末尾 overlap 字符对应的 piece
			var tail []piece
			tailSize := 0
			for i := len(buf) - 1; i >= 0; i-- {
				ts := len([]rune(buf[i].text))
				if tailSize+ts > overlap {
					break
				}
				tail = append([]piece{buf[i]}, tail...)
				tailSize += ts
			}
			buf = tail
			bufSize = tailSize
		}
		buf = append(buf, p)
		bufSize += pSize
	}
	flush()
	return chunks
}

// ──────────────────────────────────────────
// TokenSplitter
// ──────────────────────────────────────────

type tokenSplitter struct {
	opts options
}

// NewTokenSplitter 按估算 Token 数分块.
// 估算规则：ASCII 字符 4 字符 ≈ 1 token；CJK 字符 1.5 字符 ≈ 1 token.
func NewTokenSplitter(opts ...Option) Splitter {
	return &tokenSplitter{opts: applyOptions(opts)}
}

func (s *tokenSplitter) Split(text string) []Chunk {
	if text == "" {
		return nil
	}
	runes := []rune(text)
	// 将 chunkSize / chunkOverlap（字符数）换算为对应的 token 数阈值
	// 实现上仍以 rune 游标操作，但用 estimateTokens 判断窗口大小
	maxTokens := s.opts.chunkSize
	overlapTokens := s.opts.chunkOverlap
	if overlapTokens >= maxTokens {
		overlapTokens = maxTokens / 2
	}

	var chunks []Chunk
	start := 0
	for start < len(runes) {
		tokens := 0
		end := start
		for end < len(runes) {
			tokens += tokenCost(runes[end])
			end++
			// token 成本以 *10 倍定点数计算，阈值也乘以 10
			if tokens >= maxTokens*10 {
				break
			}
		}
		byteOffset := len(string(runes[:start]))
		chunks = append(chunks, Chunk{
			Text:   string(runes[start:end]),
			Offset: byteOffset,
			Index:  len(chunks),
		})
		if end == len(runes) {
			break
		}
		// 回退 overlapTokens 个 token 以形成重叠
		overlapStart := end
		ot := 0
		for overlapStart > start+1 {
			overlapStart--
			ot += tokenCost(runes[overlapStart])
			if ot >= overlapTokens*10 {
				break
			}
		}
		start = overlapStart
	}
	return chunks
}

// tokenCost 返回单个 rune 的 token 成本（以定点数 *10 表示）.
// CJK 字符：1 rune ≈ 1/1.5 token，即成本 ≈ 6.67，取整为 7.
// ASCII 字符：1 rune ≈ 1/4 token，即成本 ≈ 2.5，取整为 3.
// 其余 Unicode 字符按 CJK 处理.
func tokenCost(r rune) int {
	if r < 128 {
		// ASCII
		return 3 // 约 0.25 token * 10 = 2.5，取 3
	}
	if isCJK(r) {
		return 7 // 约 0.667 token * 10 = 6.67，取 7
	}
	// 其余多字节字符（拉丁扩展、阿拉伯语等）按 CJK 成本
	return 7
}

// isCJK 判断字符是否属于中日韩统一表意文字区块.
func isCJK(r rune) bool {
	return unicode.In(r,
		unicode.Han,
		unicode.Hiragana,
		unicode.Katakana,
		unicode.Hangul,
	)
}
