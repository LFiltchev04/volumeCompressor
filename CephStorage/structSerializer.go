package CephStorage
import(
	"io"
	"bytes"
	"encoding/gob"
)


func  serializeTarIndexArr(wc io.WriteCloser, indexArr []tarIndex) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(indexArr); err != nil {
		return err
	}
	if _, err := wc.Write(buf.Bytes()); err != nil {
		return err
	}

	return nil
}