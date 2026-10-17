package CephStorage
import(
	"bytes"
	"encoding/binary"
	"os"
	"fmt"
)


//wrt path is the globalPVs blob symlink address
func (cm *CephMount) serializeHashIndex(indexArr []xxHashStruct, wrtPath string) error {
	println("Serializing hash index to: ", wrtPath)
	buf := new(bytes.Buffer)

	file, err := cm.CephMount.Open(wrtPath, os.O_WRONLY | os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, entry := range indexArr {
		buf.Reset()
	
		binary.Write(buf, binary.LittleEndian, entry.xxHashID)
		binary.Write(buf, binary.LittleEndian, entry.recordLen)
		byteArr := []byte(entry.pathName)
		binary.Write(buf, binary.LittleEndian, byteArr)
		
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