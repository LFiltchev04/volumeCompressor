package CephStorage
import(
	"io"
)

type countingWriter struct{
	writer io.WriteCloser
	c uint64
}



func (c *countingWriter) Write(p []byte) (n int, err error) {
	n, err = c.writer.Write(p)
	c.c += uint64(n)
	return n, err
}

func (c *countingWriter) Close() error {
	return c.writer.Close()
}

func (c *countingWriter) GetCount() uint64 {
	return c.c
}