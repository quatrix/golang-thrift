package thrift

type TBufferedTransportFactory struct {
	size int
}

type TBuffer struct {
	buffer     []byte
	pos, limit int
}

type TBufferedTransport struct {
	tp   TTransport
	rbuf *TBuffer
	wbuf *TBuffer
}

func (p *TBufferedTransportFactory) GetTransport(trans TTransport) TTransport {
	return NewTBufferedTransport(trans, p.size)
}

func NewTBufferedTransportFactory(bufferSize int) *TBufferedTransportFactory {
	return &TBufferedTransportFactory{size: bufferSize}
}

func NewTBufferedTransport(trans TTransport, bufferSize int) *TBufferedTransport {
	rb := &TBuffer{buffer: make([]byte, bufferSize)}
	wb := &TBuffer{buffer: make([]byte, bufferSize), limit: bufferSize}
	return &TBufferedTransport{tp: trans, rbuf: rb, wbuf: wb}
}

func (p *TBufferedTransport) IsOpen() bool {
	return p.tp.IsOpen()
}

func (p *TBufferedTransport) Open() (err error) {
	return p.tp.Open()
}

func (p *TBufferedTransport) Close() (err error) {
	return p.tp.Close()
}

func (p *TBufferedTransport) Read(buf []byte) (n int, err error) {
	rbuf := p.rbuf
	if rbuf.pos == rbuf.limit { // no more data to read from buffer
		rbuf.pos = 0
		// read data, fill buffer
		rbuf.limit, err = p.tp.Read(rbuf.buffer)
		if err != nil {
			return 0, err
		}
	}
	n = copy(buf, rbuf.buffer[rbuf.pos:rbuf.limit])
	rbuf.pos += n
	return n, nil
}

func (p *TBufferedTransport) Write(buf []byte) (n int, err error) {
	wbuf := p.wbuf
	size := len(buf)
	if wbuf.pos+size > wbuf.limit { // buffer is full, flush buffer
		p.Flush()
	}
	n = copy(wbuf.buffer[wbuf.pos:], buf)
	wbuf.pos += n
	return n, nil
}

func (p *TBufferedTransport) Flush() error {
	start := 0
	wbuf := p.wbuf
	for start < wbuf.pos {
		n, err := p.tp.Write(wbuf.buffer[start:wbuf.pos])
		if err != nil {
			return err
		}
		start += n
	}

	wbuf.pos = 0
	return p.tp.Flush()
}

func (p *TBufferedTransport) Peek() bool {
	return p.rbuf.pos < p.rbuf.limit || p.tp.Peek()
}
