package CephStorage

import (
	"bytes"
	"encoding/json"
	"log"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	Config "compressor/Configuration"

	digest "github.com/opencontainers/go-digest"
)

//needs to have an already existing base image
func (cm *CephMount) MakeRepo(name string, baseImage string)error{
	pvBase := Config.Global.BaseConf.DistributionPV

	//get the base image and parse it
	baseJsonObj := manifestRevisionLinkPathSpec{
		name: baseImage,
		revision: "latest", //hardcodes the version pointer spec
	}
	
	baseJsonPathR, err := PathFor(baseJsonObj)
	if err != nil {
		log.Printf("Errored out on path generation %s", err.Error())
		return err
	}
	log.Println("Got base path", baseJsonPathR)


	//this is just for
	baseJsonPathA := pvBase + baseImage //no slasl? I get why people use objects for this stuff now huh
	log.Println("Got alt path", baseJsonPathA)

	jsonFilePtr := baseJsonPathA + "/link"
	rdr, err := cm.CephRead(jsonFilePtr)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(rdr)
	log.Println("Read link", buf.String())
	linkString := buf.String()
	


	dgst, err := digest.Parse(linkString)
	if err != nil {
		log.Printf("Digest parse of link string failed %s\n", err.Error())
	}
	
	manifestJsonLink := layerLinkPathSpec{
		name: "eh?",
		digest: dgst,

	}
	relPathJson, err := PathFor(manifestJsonLink)
	if err != nil {
		log.Printf("Couldnt parse the lookup string for the json %s", err.Error())
		return err
	}
	log.Println("Got JSON path", relPathJson)

	fullPath := pvBase + "/" + relPathJson
	buffer := new(bytes.Buffer)
	rdr, err = cm.CephRead(fullPath)
	if err != nil {
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
		revision: "v0",
	}
	relPathNew, err := PathFor(newRepoPath)
	if err != nil {
		return err
	}
	log.Println("New repo path", relPathNew)

	fullNewPath := pvBase + "/" + relPathNew
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
	_, err = wrtr.Write([]byte(linkString))
	if err != nil {
		return err
	}
	log.Println("Done writing link")


	//that is basically the whole thing? 
	return nil
}