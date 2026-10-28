package CephStorage

import (
	cnfg "compressor/Configuration"
	"fmt"
	"io"
	"os"
	"encoding/json"

	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)


func (cm *CephMount) LoadManifest(repoName string)(*ocispec.Manifest, error){
	
	
	rPath, err := PathFor(manifestRevisionsPathSpec{name: repoName})
	if err != nil {
		return nil, err
	}
	
	
	rPath += "/sha256/"
	fPath := cnfg.Global.BaseConf.DistributionPV + rPath
	println("Opening path:", fPath)
	revisionDir, err := cm.CephMount.OpenDir(fPath)
	if err != nil {
		return nil, err
	}
	
	//in a single manifest image uploads there ought to be only one hash name that matches completely with the manifest link content
	var manifestHash string
	for {
		entry, err := revisionDir.ReadDir()
		if err != nil {
			println("Failed to read revisions directory")
			return nil, err
		}
		if entry == nil {
			break
		}
		if entry.Name() == "." || entry.Name() == ".." {
			continue
		}
		println("Revision directory entry:", entry.Name())
		manifestHash = entry.Name()
	}


	if manifestHash == "" {
		return nil, fmt.Errorf("Error on iterating the repo dir")
	}

	dgst, err := digest.Parse(manifestHash)
	if err != nil {
		return nil, err
	}
	
	bPath := blobDataPathSpec{
		digest: dgst,
	}
	rPath, err = PathFor(bPath)

	fPath = cnfg.Global.BaseConf.DistributionPV + rPath

	file, err := cm.CephMount.Open(fPath, os.O_RDONLY, 0644)
	if err != nil {
		println("Failed to open manifest file")
		return nil, err
	}
	defer file.Close()


	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	manifest := &ocispec.Manifest{}
	if err := json.Unmarshal(data, manifest); err != nil {
		return nil, err
	}

	return manifest, nil
}
