package CephStorage

import (
	"crypto/sha256"
	"encoding/hex"
	"os"

	"compressor/Configuration"
	"path"

	digest "github.com/opencontainers/go-digest"

)

//writes a blob entry and links it independently, not recomended to stream bulk data from here
func (cm *CephMount) WriteBlob(data []byte, repoName string, newVersion string) (string, error) {
	basePath := Configuration.Global.BaseConf.DistributionPV
	

	hash := sha256.Sum256(data)
	dgst, _ := digest.Parse("sha256:" + hex.EncodeToString(hash[:]))

	relPath, _ := PathFor(blobDataPathSpec{
		digest: dgst,
	})
	fullPath := basePath + "/" + relPath
	
	dirPath := path.Dir(fullPath)
	cm.ensureDir(dirPath)
	

	file, err := cm.CephMount.Open(fullPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()

	//this can take a while for large blobs
	_, err = file.Write(data)
	if err != nil {
		return "", err
	}


	//this is only for the linkfile to go with the blobs

	linkFilePath, _ := PathFor(layerLinkPathSpec{
		name: repoName, 
		digest: dgst,
	})
	println("security check on the name stuff, the unknown nameEXT property is set here, the full path was: ", linkFilePath)
	fullLinkfilePath := basePath + "/" + linkFilePath

	dirPathL := path.Dir(fullLinkfilePath)
	cm.ensureDir(dirPathL)

	fileL, err := cm.CephMount.Open(fullLinkfilePath, os.O_CREATE|os.O_WRONLY | os.O_TRUNC, 0644)
	if err != nil {
		println("link file open err: ", err.Error())
		return "", err
	}
	defer fileL.Close()

	_, err = fileL.Write([]byte("sha256:"+ hex.EncodeToString(hash[:])))
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash[:]), nil
}