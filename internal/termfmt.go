package internal

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
)

// Formatter provides simple, consistent, colored terminal output helpers.
// Use NewFormatter to bind to a specific writer (defaults to color.Output).
// Colors are best-effort and will degrade gracefully on non-TTY.

type Formatter struct {
	Out       io.Writer
	KeyWidth  int // padding width for KV keys
	Indent    int // spaces before items
}

// NewFormatter creates a formatter writing to w (or color.Output when nil).
func NewFormatter(w io.Writer) *Formatter {
	if w == nil {
		w = color.Output
	}
	return &Formatter{Out: w, KeyWidth: 14, Indent: 0}
}

func (f *Formatter) withIndent(extra int) *Formatter {
	g := *f
	g.Indent += extra
	return &g
}

// Indent returns a new formatter with additional indentation (spaces).
func (f *Formatter) IndentBy(spaces int) *Formatter { return f.withIndent(spaces) }

func (f *Formatter) pad() string { return strings.Repeat(" ", f.Indent) }

func (f *Formatter) Header(title string) {
	bold := color.New(color.Bold).SprintFunc()
	fmt.Fprintf(f.Out, "%s%s\n", f.pad(), bold(title))
}

func (f *Formatter) Subheader(title string) {
	fmt.Fprintf(f.Out, "%s%s\n", f.pad(), color.CyanString(title))
}

func (f *Formatter) Info(format string, a ...any) {
	fmt.Fprintf(f.Out, "%s%s\n", f.pad(), fmt.Sprintf(format, a...))
}

func (f *Formatter) Success(format string, a ...any) {
	fmt.Fprintf(f.Out, "%s%s\n", f.pad(), color.GreenString(fmt.Sprintf(format, a...)))
}

func (f *Formatter) Warn(format string, a ...any) {
	fmt.Fprintf(f.Out, "%s%s\n", f.pad(), color.YellowString(fmt.Sprintf(format, a...)))
}

func (f *Formatter) Error(format string, a ...any) {
	fmt.Fprintf(f.Out, "%s%s\n", f.pad(), color.RedString(fmt.Sprintf(format, a...)))
}

func (f *Formatter) Bullet(text string) {
	fmt.Fprintf(f.Out, "%s• %s\n", f.pad(), text)
}

// KV prints aligned key/value pairs like:
//   Key:        Value
//   Longer key: Other
func (f *Formatter) KV(key string, value any) {
	k := fmt.Sprintf("%-*s", f.KeyWidth, key+":")
	fmt.Fprintf(f.Out, "%s%s %v\n", f.pad(), color.HiBlackString(k), value)
}

// PrintIf prints a line only when cond is true.
func (f *Formatter) PrintIf(cond bool, format string, a ...any) {
	if cond {
		f.Info(format, a...)
	}
}

// Global convenience helpers using a default formatter bound to stdout.
var defaultFmt = NewFormatter(nil)

func Fmt() *Formatter { return defaultFmt }

// DefaultWriter exposes the writer used by default for tests.
func DefaultWriter() io.Writer { return color.Output }

// QuietFormatter returns a formatter that discards output when quiet=true.
// When quiet is false, it returns a normal formatter to w.
func QuietFormatter(quiet bool, w io.Writer) *Formatter {
	if quiet {
		return NewFormatter(io.Discard)
	}
	if w == nil {
		w = os.Stdout
	}
	return NewFormatter(w)
}
