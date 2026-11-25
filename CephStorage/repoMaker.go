package CephStorage

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"path"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	Config "compressor/Configuration"

	digest "github.com/opencontainers/go-digest"
)

// needs to have an already existing base image
func (cm *CephMount) MakeRepo(name string, baseImage string, tagName string, pvPath string) error {
	pvBase := Config.Global.BaseConf.DistributionPV

	//get the base image and parse it
	baseJsonObj := manifestTagIndexPathSpec{
		name: baseImage,
		tag:  tagName,
	}

	baseRepoPath, err := PathFor(baseJsonObj)
	baseRepoPath += "/sha256"
	if err != nil {
		log.Printf("Errored out on path generation %s", err.Error())
		return err
	}
	log.Println("Got base path", baseRepoPath)

	//must have one repo entry only
	baseLinkPath := pvBase + baseRepoPath
	tagDir, err := cm.CephMount.OpenDir(baseLinkPath)
	if err != nil {
		log.Printf("Failed to open revisions directory %q: %s", baseLinkPath, err)
		return err
	}
	defer tagDir.Close()

	var manifestHash string
	for {
		entry, err := tagDir.ReadDir()
		if err != nil {
			log.Printf("Failed to read revisions directory %q: %s", baseLinkPath, err)
			return err
		}
		if entry == nil {
			break
		}
		if entry.Name() == "." || entry.Name() == ".." {
			continue
		}
		log.Println("Revision directory entry:", entry.Name())
		manifestHash = entry.Name()
	}

	dgst, err := digest.Parse("sha256:" + manifestHash)
	if err != nil {
		log.Printf("Digest parse of link string failed %s\n", err.Error())
		return err
	}

	//gets the blob of the json link directly
	manifestJsonLink := blobDataPathSpec{
		digest: dgst,
	}

	relPathJson, err := PathFor(manifestJsonLink)
	if err != nil {
		log.Printf("Couldnt parse the lookup string for the json %s", err.Error())
		return err
	}
	log.Println("Got JSON path", relPathJson)

	fullPath := pvBase + relPathJson
	buffer := new(bytes.Buffer)
	rdr, err := cm.CephRead(fullPath)
	if err != nil {
		log.Printf("full json path: %s", fullPath)
		log.Printf("couldnt read the json blob with error: %s", err.Error())
		return err
	}
	buffer.ReadFrom(rdr)

	//loads the new one to insert the PV path as last place entry
	mFest := ocispec.Manifest{}
	err = json.Unmarshal(buffer.Bytes(), &mFest)
	if err != nil {
		log.Printf("json unmarshal failed %s", err.Error())
		return err
	}

	tgtMap := mFest.Layers[len(mFest.Layers)-1].Annotations

	//that could maybe throw the snapshotter but IDC, the meta isnt used and i can pass garbage values so long as i forget to ditch it later
	_, exists := tgtMap["containerd.io/snapshot/volPath/"]
	if !exists {
		println("volPath not set")
		tgtMap["containerd.io/snapshot/volPath/"] = pvPath
	}


	mFest.Layers[len(mFest.Layers)-1].Annotations = map[string]string{"containerd.io/snapshot/volPath/": pvPath}
	log.Println("Manifest loaded")

	marshalledMfest, err := json.Marshal(mFest)
	if err != nil {
		log.Printf("json marshal failed %s", err.Error())
		return err
	}

	newDgst := digest.FromBytes(marshalledMfest)
	println("new digest is: ", newDgst.String())

	//write the new manifest blob to Ceph storage
	blobSpec := blobDataPathSpec{
		digest: newDgst,
	}
	relPathBlob, err := PathFor(blobSpec)
	if err != nil {
		log.Printf("Failed to generate blob path: %s", err.Error())
		return err
	}
	fullBlobPath := pvBase + relPathBlob
	err = cm.CephMount.MakeDirs(path.Dir(fullBlobPath), 0766)
	if err != nil {
		log.Printf("Failed to make blob directory: %s", err.Error())
		return err
	}

	file, err := cm.CephMount.Open(fullBlobPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Printf("Failed to open blob file: %s", err.Error())
		return err
	}
	defer file.Close()
	
	_, err = file.Write(marshalledMfest)
	if err != nil {
		log.Printf("Failed to write blob: %s", err.Error())
		return err
	}
	log.Println("Wrote new manifest blob")
	//this now appends the manifest to the new repo
	newRepoPath := manifestRevisionLinkPathSpec{
		name:     name,
		revision: newDgst,
	}
	relPathNew, err := PathFor(newRepoPath)
	if err != nil {
		return err
	}
	log.Println("New repo path", relPathNew)

	fullNewPath := pvBase + path.Dir(relPathNew)
	log.Println("Full new path", fullNewPath)

	err = cm.CephMount.MakeDirs(fullNewPath, 0766)
	if err != nil {
		log.Printf("Failed to make dirs with err: %s", err.Error())
		return err
	}

	wrtr, err := cm.CephWrite(fullNewPath + "/link")
	if err != nil {
		log.Printf("Failed to write link %s", err.Error())
		return err
	}

	log.Println("Wrote link")
	_, err = wrtr.Write([]byte(newDgst.String()))
	if err != nil {
		return err
	}
	log.Println("Done writing link")

	//now writes index/currents for default tag v0

	tagElem := manifestTagIndexEntryPathSpec{
		name:     name,
		tag:      "v0",
		revision: newDgst,
	}

	relPathTag, err := PathFor(tagElem)
	if err != nil {
		return err
	}

	fullTagDir := pvBase + relPathTag
	fullTagPath := path.Join(fullTagDir, "link")
	log.Println("Full tag path", fullTagPath)
	err = cm.CephMount.MakeDirs(fullTagDir, 0766)
	if err != nil {
		return err
	}

	file, err = cm.CephMount.Open(fullTagPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0766)
	if err != nil {
		log.Printf("index open failed %s", err.Error())
		return err
	}
	defer file.Close()

	//wrote to index
	_, err = file.Write([]byte(newDgst.String()))
	if err != nil {
		log.Printf("index write failed %s", err.Error())
		return err
	}

	//writes to current
	currElem := manifestTagCurrentPathSpec{
		name: name,
		tag:  "v0",
	}
	relPathCurr, err := PathFor(currElem)
	if err != nil {
		return err
	}
	fullCurrPath := pvBase + relPathCurr
	cm.CephMount.MakeDirs(path.Dir(fullCurrPath), 0766)
	file, err = cm.CephMount.Open(fullCurrPath, os.O_CREATE|os.O_WRONLY, 0766)
	if err != nil {
		log.Printf("current open failed %s", err.Error())
		return err
	}

	log.Println("Wrote current link")
	_, err = file.Write([]byte(newDgst.String()))
	if err != nil {
		return err
	}

	//gotta write layers

	log.Printf("Writing %d layers", len(mFest.Layers))
	for _, layer := range mFest.Layers {
		log.Println("Layer", layer.Digest)

		lyrElem := layerLinkPathSpec{
			name:   name,
			digest: layer.Digest,
		}
		relPathLyr, err := PathFor(lyrElem)
		if err != nil {
			return err
		}
		fLyrPath := pvBase + relPathLyr

		log.Println("Layer path", fLyrPath)

		fLyrDir := path.Dir(fLyrPath)
		err = cm.ensureDir(fLyrDir)
		if err != nil {
			return err
		}

		file, err := cm.CephMount.Open(fLyrDir+"/link", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0766)
		if err != nil {
			return err
		}

		_, err = file.Write([]byte(layer.Digest))
		if err != nil {
			return err
		}
		log.Println("Wrote layer: ", layer.Digest)
		file.Close()
	}

	//add the config json for good measure
	confElem := layerLinkPathSpec{
		name:   name,
		digest: mFest.Config.Digest,
	}
	relPathConf, _ := PathFor(confElem)
	cm.ensureDir(pvBase + relPathConf)
	file, _ = cm.CephMount.Open(pvBase+relPathConf+"/link", os.O_CREATE|os.O_WRONLY, 0766)
	file.Write([]byte(mFest.Config.Digest))

	//that is basically the whole thing?
	return nil
}
