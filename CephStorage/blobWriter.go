package CephStorage

import (
	"crypto/sha256"
	"os"

	"compressor/Configuration"

	digest "github.com/opencontainers/go-digest"

)

//writes a blob entry and links it independently, not recomended to stream bulk data from here
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

	defer file.Close()


	//this is only for the linkfile to go with the blobs

	linkFilePath, _ := PathFor(layerLinkPathSpec{
		name: "linkEXT", //i dont know why this exists, it is never asked for
		digest: dgst,
	})
	println("security check on the name stuff, the unknown nameEXT property is set here, the full path was: ", linkFilePath)
	fullLinkfilePath := basePath + "/" + linkFilePath


	fileL, err := cm.CephMount.Open(fullLinkfilePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		defer fileL.Close()
		return err
	}

	_, err = fileL.Write([]byte{})
	if err != nil {
		defer fileL.Close()
		return err
	}

	defer fileL.Close()
	return nil
}