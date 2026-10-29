package CephStorage

import (
	cnfg "compressor/Configuration"
	"os"
	digest "github.com/opencontainers/go-digest"
)



//this is a metadata only operation for making an already existing image visible to the file server for pulling 
func (cm* CephMount) WriteToLayers(repoName string, tagName string, hash string) error {

	dgst, err := digest.Parse(hash)
	if err != nil {
		return err
	}


	manifestLink := layerLinkPathSpec{
		name: repoName,
		digest: dgst,
	}
	layerWriter, err := PathFor(manifestLink)
	if err != nil {
		println("Error getting layer path")
		return err
	}

	rLyrPath := cnfg.Global.BaseConf.DistributionPV + layerWriter

	file, err := cm.CephMount.Open(rLyrPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 644)
	if err != nil {
		println("Error opening layer file")
		return err
	}


	println("Writing to layer")
	file.Write([]byte(dgst.String()))
	file.Close()

	println("Layer write done")

	//should be completely readable by now, have to advance tag so current works, optionally


	return nil
}