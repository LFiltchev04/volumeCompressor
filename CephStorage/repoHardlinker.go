package CephStorage

import (
	"compressor/Configuration"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path"

	//"os"
	"fmt"

	"github.com/ceph/go-ceph/cephfs"
	"github.com/opencontainers/go-digest"
)

//this method is likely pretty ineffective, will do a second trip to the repo to fetch it out and hash it adter a write, have to look at alternatives
func (cm* CephMount)blobHasher(fileLocation string)(string, error){
	println("path is: ", fileLocation)
	file, err := cm.CephMount.Open(fileLocation,0,0644)
	if err != nil {
		println("errored out on raw bob hasher open with error: ", err.Error())
	}
	
	fSize, _ := file.Fstatx(cephfs.StatxSize,0)

	hasher := sha256.New()
	println("about to fork over: ", fSize.Size, " bytes")
	io.CopyN(hasher, file, int64(fSize.Size))
	
	byteDump := hasher.Sum(nil)
	println("hash is: ", hex.EncodeToString(byteDump))
	return fmt.Sprintf("%x", byteDump), nil
}


//hardlinks a pointer in the cncf distribution CAS to a real file in the pvc backend
//shouldve called it symlinker since it aint a hardlink but i dont want to go digging here
func (cm* CephMount)HardlinkBlob(blobPath string, repoName string, vname string)error{

	hash, err := cm.blobHasher(blobPath)
	if err != nil {
		println("err on hasher: ", err.Error())
		return err
	}
	println("finished hash")

	dgst, _ := digest.Parse(hash)
	println("dgst created", dgst.String())
	pth, err := PathFor(layerLinkPathSpec{
		name: repoName,
		digest: dgst,
	})
	if err != nil {
		return err
	}



	linkPath := Configuration.Global.BaseConf.DistributionPV + pth
	pDir := path.Dir(pth)

	cm.CephMount.MakeDirs(pDir,0644)
	println("The. dir made is: ", pDir)

	cm.CephMount.Open(linkPath,os.O_CREATE,0644)



	writeHandle, err := cm.CephWrite(linkPath)
	if err != nil {
		println("error writing linkfile: ", err.Error())
		return  err
	}


	hash = dgst.String()

	linkContent := hash
	println("The link content is: ", linkContent)
	_, err = writeHandle.Write([]byte(linkContent))
	if err != nil {
		println("error writing link: ", err.Error())
		return err
	}
	
	
	pth, _ = PathFor(blobDataPathSpec{
		digest: dgst,
	})

	hardlinkDestination := Configuration.Global.BaseConf.DistributionPV + pth

	pDirHardl := path.Dir(hardlinkDestination)
	cm.CephMount.MakeDirs(pDirHardl, 0644)

	println("pth is: ", hardlinkDestination, " for blob path: ", blobPath)
	errV := cm.CephMount.Symlink(blobPath,hardlinkDestination)
	if errV != nil {
		println("error on symlinking is: ", errV.Error())
		return err
	}

	return nil
}

