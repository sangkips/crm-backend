package printer

import (
	"bytes"
	"fmt"
	"strings"
)

// ESC/POS command constants
const (
	ESC = 0x1B
	GS  = 0x1D
	LF  = 0x0A
)

// Text alignment
const (
	AlignLeft   = 0
	AlignCenter = 1
	AlignRight  = 2
)

// Font size
const (
	FontNormal = 0x00
	FontDouble = 0x11 // Double width + double height
	FontWide   = 0x10 // Double width only
	FontTall   = 0x01 // Double height only
)

// Document builds an ESC/POS byte stream for thermal printers.
type Document struct {
	buf   bytes.Buffer
	width int // print width in characters (default 32 for 58mm, 48 for 80mm)
}

// NewDocument creates a new ESC/POS document with the given character width.
// Common widths: 32 for 58mm paper, 48 for 80mm paper.
func NewDocument(charWidth int) *Document {
	if charWidth <= 0 {
		charWidth = 32
	}
	d := &Document{width: charWidth}
	d.Init()
	return d
}

// Init sends the ESC @ (initialize printer) command.
func (d *Document) Init() *Document {
	d.buf.Write([]byte{ESC, '@'})
	return d
}

// LineFeed sends a line feed.
func (d *Document) LineFeed() *Document {
	d.buf.WriteByte(LF)
	return d
}

// FeedLines sends n line feeds.
func (d *Document) FeedLines(n int) *Document {
	for i := 0; i < n; i++ {
		d.buf.WriteByte(LF)
	}
	return d
}

// SetAlign sets text alignment: AlignLeft, AlignCenter, AlignRight.
func (d *Document) SetAlign(align int) *Document {
	d.buf.Write([]byte{ESC, 'a', byte(align)})
	return d
}

// SetBold enables or disables bold text.
func (d *Document) SetBold(on bool) *Document {
	b := byte(0)
	if on {
		b = 1
	}
	d.buf.Write([]byte{ESC, 'E', b})
	return d
}

// SetFontSize sets the character size. Use FontNormal, FontDouble, FontWide, or FontTall.
func (d *Document) SetFontSize(size byte) *Document {
	d.buf.Write([]byte{GS, '!', size})
	return d
}

// Text writes a line of text followed by a line feed.
func (d *Document) Text(s string) *Document {
	d.buf.WriteString(s)
	d.buf.WriteByte(LF)
	return d
}

// TextF writes a formatted line of text followed by a line feed.
func (d *Document) TextF(format string, args ...interface{}) *Document {
	d.buf.WriteString(fmt.Sprintf(format, args...))
	d.buf.WriteByte(LF)
	return d
}

// Separator prints a full-width separator line (e.g. "--------------------------------").
func (d *Document) Separator(char byte) *Document {
	d.buf.WriteString(strings.Repeat(string(char), d.width))
	d.buf.WriteByte(LF)
	return d
}

// KeyValue prints a left-aligned key and right-aligned value on the same line.
// Example: "Subtotal           $100.00"
func (d *Document) KeyValue(key, value string) *Document {
	spaces := d.width - len(key) - len(value)
	if spaces < 1 {
		spaces = 1
	}
	d.buf.WriteString(key)
	d.buf.WriteString(strings.Repeat(" ", spaces))
	d.buf.WriteString(value)
	d.buf.WriteByte(LF)
	return d
}

// ItemLine prints a receipt item line: qty x name, then right-aligned total.
// Example: "2x Widget              $20.00"
func (d *Document) ItemLine(qty int, name, total string) *Document {
	prefix := fmt.Sprintf("%dx %s", qty, name)
	spaces := d.width - len(prefix) - len(total)
	if spaces < 1 {
		spaces = 1
	}
	d.buf.WriteString(prefix)
	d.buf.WriteString(strings.Repeat(" ", spaces))
	d.buf.WriteString(total)
	d.buf.WriteByte(LF)
	return d
}

// Cut sends the paper cut command (full cut).
func (d *Document) Cut() *Document {
	d.buf.Write([]byte{GS, 'V', 0x00})
	return d
}

// PartialCut sends the partial cut command.
func (d *Document) PartialCut() *Document {
	d.buf.Write([]byte{GS, 'V', 0x01})
	return d
}

// Bytes returns the accumulated ESC/POS byte stream.
func (d *Document) Bytes() []byte {
	return d.buf.Bytes()
}

// Reset clears the buffer and reinitializes the document.
func (d *Document) Reset() *Document {
	d.buf.Reset()
	d.Init()
	return d
}
