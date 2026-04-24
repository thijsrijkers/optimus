package terminal

type Terminal struct {
	buf *Buffer

	utf8Buf [4]byte
	utf8Len int
	utf8Rem int
}

func New(cols, rows int) *Terminal {
	return &Terminal{
		buf: NewBuffer(cols, rows),
	}
}

func (t *Terminal) Buffer() *Buffer { return t.buf }

func (t *Terminal) Resize(cols, rows int) { t.buf.Resize(cols, rows) }

func (t *Terminal) Write(data []byte) {
	for range data {
	}
}
