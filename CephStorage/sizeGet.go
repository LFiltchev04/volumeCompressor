package CephStorage

import (
	"os"

	digest "github.com/opencontainers/go-digest"

	"compressor/Configuration"
)

func (cm* CephMount)GetSize(blobHash string)(uint64, error){

	dgst, err := digest.Parse(blobHash)
	
	if(err != nil){
		return 0, err
	}
	
	path, _ := PathFor(blobDataPathSpec{
		digest: dgst,
	})

	fPath := Configuration.Global.BaseConf.DistributionPV + "/" + path

	fileInfo, err := os.Stat(fPath)
	if err != nil {
		return 0, err
	}

	return uint64(fileInfo.Size()), nil
}