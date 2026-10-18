package CephStorage

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

// wrt path is the globalPVs blob symlink address
func (cm *CephMount) serializeHashIndex(indexArr []xxHashStruct, wrtPath string) error {
	println("Serializing hash index to: ", wrtPath)
	buf := new(bytes.Buffer)

	file, err := cm.CephMount.Open(wrtPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, entry := range indexArr {
		buf.Reset()

		for _, field := range []any{
			entry.xxHashID,
			entry.fileSize,
			entry.offsetPos,
			entry.recordLen,
		} {
			if err := binary.Write(buf, binary.LittleEndian, field); err != nil {
				return fmt.Errorf("serialize hash index entry: %w", err)
			}
		}
		if entry.recordLen != uint64(len(entry.pathName)) {
			return fmt.Errorf("path length mismatch for hash index entry: recorded %d bytes, got %d", entry.recordLen, len(entry.pathName))
		}
		if _, err := buf.WriteString(entry.pathName); err != nil {
			return fmt.Errorf("serialize hash index path: %w", err)
		}

		num, err := file.Write(buf.Bytes())
		if err != nil {
			return err
		}
		if num != buf.Len() {
			return fmt.Errorf("incomplete write: wrote %d bytes, expected %d", num, buf.Len())
		}
	}
	return nil
}
