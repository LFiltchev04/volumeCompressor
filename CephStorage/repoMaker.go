package CephStorage

import (
	"bytes"
	"encoding/json"
	"log"
	"path"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	Config "compressor/Configuration"

	digest "github.com/opencontainers/go-digest"
)

//needs to have an already existing base image
func (cm *CephMount) MakeRepo(name string, baseImage string)error{
	pvBase := Config.Global.BaseConf.DistributionPV

	//get the base image and parse it
	baseJsonObj := manifestRevisionsPathSpec{
		name: baseImage,
	}
	
	baseRepoPath, err := PathFor(baseJsonObj)
	if err != nil {
		log.Printf("Errored out on path generation %s", err.Error())
		return err
	}
	log.Println("Got base path", baseRepoPath)


	//must have one repo entry only
	baseLinkPath := pvBase + baseRepoPath
	revisionDir, err := cm.CephMount.OpenDir(baseLinkPath + "/sha256")
	if err != nil {
		log.Printf("Failed to open revisions directory %q: %s", baseLinkPath, err)
		return err
	}
	defer revisionDir.Close()

	var manifestHash string
	for {
		entry, err := revisionDir.ReadDir()
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


	dgst, err := digest.Parse("sha256:"+manifestHash)
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

	mFest := ocispec.Manifest{}
	err = json.Unmarshal(buffer.Bytes(), &mFest)
	if err != nil {
		log.Printf("json unmarshal failed %s", err.Error())
		return err
	}

	log.Println("Manifest loaded")

	

	//this now appends the manifest to the new repo
	newRepoPath := manifestRevisionLinkPathSpec{
		name: name,
		revision: dgst,
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
	_, err = wrtr.Write([]byte(dgst.String()))
	if err != nil {
		return err
	}
	log.Println("Done writing link")


	
	//that is basically the whole thing? 
	return nil
}