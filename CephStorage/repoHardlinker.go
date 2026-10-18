package CephStorage

import (
	"compressor/Configuration"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"path"
	"strings"

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
		return "", err
	}
	defer file.Close()

	fSize, err := file.Fstatx(cephfs.StatxSize,0)

	if err != nil {
		println("errored out on raw bob hasher fstatx with error: ", err.Error())
		return "", err
	}
	hasher := sha256.New()
	println("about to fork over: ", fSize.Size, " bytes")
	io.CopyN(hasher, file, int64(fSize.Size))
	
	byteDump := hasher.Sum(nil)
	println("hash is: ", hex.EncodeToString(byteDump))
	return fmt.Sprintf("%x", byteDump), nil
}


// ensureDir avoids ceph_mkdirs (which blocks here) by stat-ing first and only
// creating the components that are actually missing.
func (cm *CephMount) ensureDir(dir string) error {
	if _, err := cm.CephMount.Statx(dir, cephfs.StatxMode, 0); err == nil {
		println("dir already exists: ", dir)
		return nil
	}

	cur := ""
	for _, part := range strings.Split(strings.Trim(dir, "/"), "/") {
		cur += "/" + part
		if _, err := cm.CephMount.Statx(cur, cephfs.StatxMode, 0); err == nil {
			continue
		}
		println("creating dir: ", cur)
		if err := cm.CephMount.MakeDir(cur, 0755); err != nil {
			// a concurrent creator or a stat race is fine as long as it exists now
			if _, serr := cm.CephMount.Statx(cur, cephfs.StatxMode, 0); serr != nil {
				return fmt.Errorf("mkdir %q: %w", cur, err)
			}
		}
	}
	return nil
}

//hardlinks a pointer in the cncf distribution CAS to a real file in the pvc backend
//shouldve called it symlinker since it aint a hardlink but i dont want to go digging here
func (cm* CephMount)HardlinkBlob(blobPath string, repoName string, vname string)error{

	//first have to actually write the blob to the blob store





	println("Hardlinking blob for path: ", blobPath)
	hash, err := cm.blobHasher(blobPath)
	if err != nil {
		println("err on hasher: ", err.Error())
		return err
	}

	println("exited hasher")
	//renames it to hash value
	cm.CephMount.Rename(blobPath, hash)

	println("finished hash")

	dgst, err := digest.Parse("sha256:" + hash)
	if err != nil {
		return fmt.Errorf("parse blob digest: %w", err)
	}
	println("dgst created", dgst.String())
	pth, err := PathFor(layerLinkPathSpec{
		name: repoName,
		digest: dgst,
	})
	if err != nil {
		println("error on PathFor: ", err.Error())
		return err
	}




	linkPath := Configuration.Global.BaseConf.DistributionPV + pth
	println("//////link path is: ", linkPath)
	pDir := path.Dir(linkPath)

	err = cm.ensureDir(pDir)
	if err != nil {
		println("error making dirs: ", err.Error())
		return err
	}
	println("The. dir made is: ", pDir)

	writeHandle, err := cm.CephWrite(linkPath)
	if err != nil {
		println("error writing linkfile: ", err.Error())
		return  err
	}
	defer writeHandle.Close()


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
	err = cm.ensureDir(pDirHardl)
	if err != nil {
		println("error making hardlink dirs: ", err.Error())
		return err
	}

	println("pth is: ", hardlinkDestination, " for blob path: ", blobPath)
	errV := cm.CephMount.Symlink(blobPath,hardlinkDestination)
	if errV != nil {
		println("error on symlinking is: ", errV.Error())
		return errV
	}

	return nil
}
