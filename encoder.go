package encgen

import (
	"encoding/json"
	"io"
	"unicode/utf8"
)

// Encoder is a struct which is embedded in all generated encoders. It provides a number of
// helper functions to simplify generated code.
//
// In order to keep this easy to use, we do not return an error at each step.
// Instead, we store the error in the `err` field, and skip future write operations
// if `err` is not nil. This allows us to keep performance in case of an error, but
// not require explicit error checks at every single step.
//
// Errors can be checked for manually using the `Error()` method.
type Encoder struct {
	w   io.Writer
	err error
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

func (e Encoder) Error() error {
	return e.err
}

func (e *Encoder) Field(name string, val any) {
	if e.err != nil {
		return
	}

	bs, err := json.Marshal(val)
	if err != nil {
		e.err = err
		return
	}

	e.Byte('"')
	e.String(name)
	e.Byte('"', ':')
	e.Byte(bs...)
}

func (e *Encoder) Byte(b ...byte) {
	if e.err != nil {
		return
	}

	_, e.err = e.w.Write(b)
}

func (e *Encoder) writeRune(r rune) {
	if e.err != nil {
		return
	}

	buf := make([]byte, utf8.RuneLen(r))
	n := utf8.EncodeRune(buf, r)
	_, e.err = e.w.Write(buf[:n])
}

func (e *Encoder) String(str string) {
	if e.err != nil {
		return
	}

	_, e.err = io.WriteString(e.w, str)
}

func (e *Encoder) Marshal(val any) {
	if e.err != nil {
		return
	}

	bs, err := json.Marshal(val)
	if err != nil {
		e.err = err
		return
	}

	e.Byte(bs...)
}

func (e *Encoder) OpenObject() {
	e.Byte('{')
}

func (e *Encoder) CloseObject() {
	e.Byte('}')
}

func (e *Encoder) OpenArray() {
	e.Byte('[')
}

func (e *Encoder) CloseArray() {
	e.Byte(']')
}

func (e *Encoder) Comma() {
	e.Byte(',')
}
