package splitter

import (
	"strings"
	"testing"
)

// repeat 生成由字符 c 重复 n 次的字符串（用于构造测试数据）.
func repeat(c string, n int) string {
	return strings.Repeat(c, n)
}

// TestCharacterSplitter 测试按字符数分块，验证块数量与重叠内容.
func TestCharacterSplitter(t *testing.T) {
	// 构造 500 个 'a' 组成的字符串
	text := repeat("a", 500)
	s := NewCharacterSplitter(WithChunkSize(200), WithChunkOverlap(50))
	chunks := s.Split(text)

	if len(chunks) == 0 {
		t.Fatal("期望得到若干块，实际为空")
	}
	// 验证每块长度不超过 chunkSize
	for i, c := range chunks {
		if len([]rune(c.Text)) > 200 {
			t.Errorf("块 %d 长度 %d 超过 chunkSize 200", i, len([]rune(c.Text)))
		}
	}
	// 验证相邻块存在重叠（不含最后一块）
	for i := 1; i < len(chunks); i++ {
		prev := chunks[i-1].Text
		cur := chunks[i].Text
		// 前一块的后缀应出现在当前块的前缀中
		suffix := prev[len(prev)-min(50, len(prev)):]
		if !strings.HasPrefix(cur, suffix) {
			t.Errorf("块 %d 与块 %d 之间缺少预期重叠", i-1, i)
		}
	}
	// 验证 Index 连续
	for i, c := range chunks {
		if c.Index != i {
			t.Errorf("块 %d 的 Index=%d，期望 %d", i, c.Index, i)
		}
	}
}

// TestRecursiveSplitter 测试递归分块，验证段落与句子级别的切分.
func TestRecursiveSplitter(t *testing.T) {
	// 构造包含段落与句子的文本
	para1 := "第一段第一句。第一段第二句。第一段第三句。"
	para2 := "第二段第一句。第二段第二句。"
	para3 := "第三段内容，这是一段比较短的文字。"
	text := para1 + "\n\n" + para2 + "\n\n" + para3

	s := NewRecursiveSplitter(WithChunkSize(30), WithChunkOverlap(5))
	chunks := s.Split(text)

	if len(chunks) == 0 {
		t.Fatal("期望得到若干块，实际为空")
	}
	// 验证所有块内容能在原文中找到
	for i, c := range chunks {
		if !strings.Contains(text, c.Text) {
			t.Errorf("块 %d 的内容 %q 不属于原文", i, c.Text)
		}
	}
	// 验证 Index 连续
	for i, c := range chunks {
		if c.Index != i {
			t.Errorf("块 %d 的 Index=%d，期望 %d", i, c.Index, i)
		}
	}
}

// TestTokenSplitter 测试按 Token 数估算分块，验证英中混合文本的切分结果.
func TestTokenSplitter(t *testing.T) {
	// 混合英文与中文
	english := strings.Repeat("hello world ", 30) // 约 360 ASCII 字符 ≈ 90 tokens
	chinese := strings.Repeat("你好世界", 20)         // 80 CJK 字符 ≈ 53 tokens
	text := english + chinese

	s := NewTokenSplitter(WithChunkSize(50), WithChunkOverlap(10))
	chunks := s.Split(text)

	if len(chunks) == 0 {
		t.Fatal("期望得到若干块，实际为空")
	}
	// 验证 Index 连续
	for i, c := range chunks {
		if c.Index != i {
			t.Errorf("块 %d 的 Index=%d，期望 %d", i, c.Index, i)
		}
	}
	// 验证所有块都是原文的子串
	for i, c := range chunks {
		if !strings.Contains(text, c.Text) {
			t.Errorf("块 %d 内容不属于原文", i)
		}
	}
}

// TestChunkOffsets 验证各块的 Offset 字段与原文中实际起始字节偏移一致.
func TestChunkOffsets(t *testing.T) {
	text := "Hello, 世界！This is a test."
	s := NewCharacterSplitter(WithChunkSize(5), WithChunkOverlap(0))
	chunks := s.Split(text)

	runes := []rune(text)
	for _, c := range chunks {
		// 根据 Offset 从原文切出相同长度的内容，应与 c.Text 一致
		chunkRunes := []rune(c.Text)
		// 将字节偏移换算回 rune 偏移
		bytesBefore := text[:c.Offset]
		runeOffset := len([]rune(bytesBefore))
		end := runeOffset + len(chunkRunes)
		if end > len(runes) {
			end = len(runes)
		}
		expected := string(runes[runeOffset:end])
		if expected != c.Text {
			t.Errorf("块 Index=%d: Offset=%d 对应文本 %q，期望 %q",
				c.Index, c.Offset, expected, c.Text)
		}
	}
}

// TestEmptyText 验证空字符串输入返回空切片.
func TestEmptyText(t *testing.T) {
	for _, s := range []Splitter{
		NewCharacterSplitter(),
		NewRecursiveSplitter(),
		NewTokenSplitter(),
	} {
		chunks := s.Split("")
		if len(chunks) != 0 {
			t.Errorf("%T: 空文本应返回空切片，实际返回 %d 块", s, len(chunks))
		}
	}
}

// TestSmallText 验证小于 chunkSize 的文本只返回单一块且内容完整.
func TestSmallText(t *testing.T) {
	text := "短文本测试"
	for _, s := range []Splitter{
		NewCharacterSplitter(),
		NewRecursiveSplitter(),
		NewTokenSplitter(),
	} {
		chunks := s.Split(text)
		if len(chunks) != 1 {
			t.Errorf("%T: 小文本应返回 1 块，实际返回 %d 块", s, len(chunks))
			continue
		}
		if chunks[0].Text != text {
			t.Errorf("%T: 块内容 %q 与原文 %q 不符", s, chunks[0].Text, text)
		}
		if chunks[0].Index != 0 {
			t.Errorf("%T: 单块 Index 应为 0，实际为 %d", s, chunks[0].Index)
		}
		if chunks[0].Offset != 0 {
			t.Errorf("%T: 单块 Offset 应为 0，实际为 %d", s, chunks[0].Offset)
		}
	}
}

// min 返回两个整数中的较小值.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
