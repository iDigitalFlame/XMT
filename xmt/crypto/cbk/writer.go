package cbk

import (
	"io"

	"github.com/iDigitalFlame/xmt/xmt/data"
)

type Writer struct {
	Cipher *Cipher

	w io.Writer
}

func NewWriter(c *Cipher, w io.Writer) *Writer {
	return &Writer{
		w:      w,
		Cipher: c,
	}
}

// Flush commits any outstanding writes to the underlying Writer and clears the writing buffer.
// This function will call the Flush function on the underlying Writer if supported.
func (w *Writer) Flush() error {
	_, err := w.Cipher.syncOutput(w)
	if f, ok := w.w.(data.Flusher); ok {
		err = f.Flush()
	}
	return err
}

// Close calls the Flush function and calls the Close function on the underlying Writer if supported.
func (w *Writer) Close() error {
	if err := w.Flush(); err != nil {
		return err
	}
	if c, ok := w.w.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

func (w *Writer) Write(b []byte) (int, error) {
	return w.Cipher.write(w.w, b)
}
