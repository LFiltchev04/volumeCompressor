package CephStorage

import(
	"crypto/sha256"
	"encoding/json"
	"os"

	"github.com/opencontainers/image-spec/specs-go/v1"
	digest "github.com/opencontainers/go-digest"
    
	"compressor/Configuration"
)


//writes a new tag for a given image and makes it the current latest tag, assumes the manifest blob already exists
func (cm *CephMount) AdvanceTag(tagName string, envName string ,manifest *v1.Manifest) error {
	rootP := Configuration.Global.BaseConf.DistributionPV

	//writes the new one to the indexed location
	jsonMfst, _ := json.Marshal(manifest)
	revisionHash := sha256.Sum256(jsonMfst)
	stringCast := string(revisionHash[:])

	revisionn, _ := digest.Parse(stringCast)

	relPath, _ := PathFor(manifestTagIndexEntryPathSpec{
		name: envName,
		tag: tagName,
		revision: revisionn,
	})

	fPath := rootP + "/" +relPath

	file, err := cm.CephMount.Open(fPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	_, err = file.Write(jsonMfst)
	if err != nil {
		return err
	}
	file.Close()


	//sets the topmost file pointer to this thing



	return nil
}