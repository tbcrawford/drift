package drift

import "io"

// Builder accumulates functional options for Diff, Render, and RenderWithNames.
// Use New to start a chain; each With* equivalent returns the same builder for chaining.
type Builder struct {
	opts []Option
}

// New returns an empty Builder. Options resolve through the same defaults as package-level Diff and Render.
func New() *Builder {
	return &Builder{}
}

// Algorithm appends WithAlgorithm(a).
func (b *Builder) Algorithm(a Algorithm) *Builder {
	b.opts = append(b.opts, WithAlgorithm(a))
	return b
}

// Context appends WithContext(n).
func (b *Builder) Context(n int) *Builder {
	b.opts = append(b.opts, WithContext(n))
	return b
}

// NoColor appends WithNoColor().
func (b *Builder) NoColor() *Builder {
	b.opts = append(b.opts, WithNoColor())
	return b
}

// Lang appends WithLang(lang).
func (b *Builder) Lang(lang string) *Builder {
	b.opts = append(b.opts, WithLang(lang))
	return b
}

// Theme appends WithTheme(theme).
func (b *Builder) Theme(theme string) *Builder {
	b.opts = append(b.opts, WithTheme(theme))
	return b
}

// Split appends WithSplit().
func (b *Builder) Split() *Builder {
	b.opts = append(b.opts, WithSplit())
	return b
}

// LineNumbers appends WithLineNumbers(v).
func (b *Builder) LineNumbers(v bool) *Builder {
	b.opts = append(b.opts, WithLineNumbers(v))
	return b
}

// WithoutLineNumbers appends WithoutLineNumbers().
func (b *Builder) WithoutLineNumbers() *Builder {
	b.opts = append(b.opts, WithoutLineNumbers())
	return b
}

// LineDiffStyle appends WithLineDiffStyle(v).
func (b *Builder) LineDiffStyle(v bool) *Builder {
	b.opts = append(b.opts, WithLineDiffStyle(v))
	return b
}

// WordDiff appends WithWordDiff(v).
func (b *Builder) WordDiff(v bool) *Builder {
	b.opts = append(b.opts, WithWordDiff(v))
	return b
}

// Diff runs the line-level diff with the accumulated options.
func (b *Builder) Diff(old, new string) (DiffResult, error) {
	return Diff(old, new, b.opts...)
}

// Render writes result to w using the accumulated options.
func (b *Builder) Render(result DiffResult, w io.Writer) error {
	return Render(result, w, b.opts...)
}

// RenderWithNames is like Render but sets file path labels in the diff header.
func (b *Builder) RenderWithNames(result DiffResult, w io.Writer, oldName, newName string) error {
	return RenderWithNames(result, w, oldName, newName, b.opts...)
}
