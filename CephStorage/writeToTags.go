package CephStorage

import (
	"crypto/sha256"
	cnfg "compressor/Configuration"
	"os"
	"path"
	digest "github.com/opencontainers/go-digest"
)




func (cm *CephMount) WriteToTags(repoName string, tag string, mFest []byte){
	
	sha256sum := sha256.Sum256(mFest)
	dgst := digest.NewDigestFromBytes(digest.SHA256, sha256sum[:])
	
	tagPath := manifestTagIndexEntryLinkPathSpec{
		name: repoName,
		tag: tag,
		revision: dgst,
	}
	rPath, err := PathFor(tagPath)
	if err != nil {
		return
	}

	fPath := cnfg.Global.BaseConf.DistributionPV + rPath
	println("This is the tag insert path", fPath)
	pDir := path.Dir(fPath)

	cm.ensureDir(pDir)
	file, err := cm.CephMount.Open(fPath, os.O_CREATE | os.O_TRUNC, 644)
	if err != nil {
		return
	}

	//has completed a write to the repo 
	println("Will write tag link: ", dgst.String())
	file.Write([]byte(dgst.String()))

	file.Close()

	println("Tag link written")

}