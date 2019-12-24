package data

import "io"

// ReadStringList attempts to read a string list written using
// the 'WriteStringList' function from the supplied string into
// the string list pointer. If the provided array is nil or not large
// enough, it will be resized.
func ReadStringList(r Reader, s *[]string) error {
	t, err := r.Uint8()
	if err != nil {
		return err
	}
	var l int
	switch t {
	case 0:
		return nil
	case 1, 2:
		n, err := r.Uint8()
		if err != nil {
			return err
		}
		l = int(n)
	case 3, 4:
		n, err := r.Uint16()
		if err != nil {
			return err
		}
		l = int(n)
	case 5, 6:
		n, err := r.Uint32()
		if err != nil {
			return err
		}
		l = int(n)
	case 7, 8:
		n, err := r.Uint64()
		if err != nil {
			return err
		}
		l = int(n)
	default:
		return ErrInvalidString
	}
	if s == nil || len(*s) < l {
		*s = make([]string, l)
	}
	for x := 0; x < l; x++ {
		if err := r.ReadString(&(*s)[x]); err != nil {
			return err
		}
	}
	return nil
}

// WriteStringList will attempt to write the supplied string list to
// the writer. If the string list is nil or empty, it will write a zero
// byte to the Writer. The resulting data can be read using the 'ReadStringList'
// function.
func WriteStringList(w Writer, s []string) error {
	if s == nil {
		return w.WriteUint8(0)
	}
	switch l := len(s); {
	case l == 0:
		return w.WriteUint8(0)
	case l < DataLimitSmall:
		if err := w.WriteUint8(1); err != nil {
			return err
		}
		if err := w.WriteUint8(uint8(l)); err != nil {
			return err
		}
	case l < DataLimitMedium:
		if err := w.WriteUint8(3); err != nil {
			return err
		}
		if err := w.WriteUint16(uint16(l)); err != nil {
			return err
		}
	case l < DataLimitLarge:
		if err := w.WriteUint8(5); err != nil {
			return err
		}
		if err := w.WriteUint32(uint32(l)); err != nil {
			return err
		}
	default:
		if err := w.WriteUint8(7); err != nil {
			return err
		}
		if err := w.WriteUint64(uint64(l)); err != nil {
			return err
		}
	}
	for i := range s {
		if err := w.WriteString(s[i]); err != nil {
			return err
		}
	}
	return nil
}

// ReadFully attempts to Read all the bytes from the
// specified reader until the length of the array or EOF.
func ReadFully(r io.Reader, b []byte) (int, error) {
	var n int
	for n < len(b) {
		i, err := r.Read(b[n:])
		if err != nil && (err != io.EOF || n != len(b)) {
			return n, err
		}
		n += i
	}
	return n, nil
}
