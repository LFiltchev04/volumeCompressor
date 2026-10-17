package CephStorage
import(
	"bytes"
	"encoding/binary"
	"os"
)


//wrt path is the globalPVs blob symlink address
func (cm *CephMount) serializeHashIndex(indexArr []xxHashStruct, wrtPath string) error {
	
	buf := new(bytes.Buffer)

	file, err := cm.CephMount.Open(wrtPath, os.O_WRONLY, 0644)
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
		
		file.Write(buf.Bytes())
	}



	return nil
}