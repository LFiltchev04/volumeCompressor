package CephStorage

import (
	"crypto/sha256"
	"os"

	"compressor/Configuration"

	digest "github.com/opencontainers/go-digest"

)

//writes a blob entry and links it independently
func (cm *CephMount) WriteBlob(data *[]byte) error {
	basePath := Configuration.Global.BaseConf.DistributionPV
	

	hash := sha256.Sum256(*data)
	dgst, _ := digest.Parse("sha256:" + string(hash[:]))

	relPath, _ := PathFor(blobDataPathSpec{
		digest: dgst,
	})
	fullPath := basePath + "/" + relPath

	file, err := cm.CephMount.Open(fullPath, os.O_CREATE|os.O_WRONLY, 0644)
	if(err!=nil){
		defer file.Close()
		return err
	}

	//this can take a while for large blobs
	_, err = file.Write(*data)
	if err != nil {
		defer file.Close()
		return err
	}


	//this is only for the linkfile to go with the blobs

	

	defer file.Close()
	return nil
}