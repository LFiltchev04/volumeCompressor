package CephStorage


import (
    specs "github.com/opencontainers/image-spec/specs-go"
	config "compressor/Configuration"
    "github.com/ceph/go-ceph/cephfs"
    veeOne "github.com/opencontainers/image-spec/specs-go/v1"
    openconcept "github.com/opencontainers/go-digest"
	"encoding/json"
)


//assumes a blob already exists
func formatManifest(layerHashes []string)string{

	baseP := config.Global.BaseConf.DistributionPV + "/docker/registry/v2/blobs/sha256"

	var annotationMap map[string]string = make(map[string]string)
	layers := []veeOne.Descriptor{}
	for _, hash := range layerHashes {

		
		hashPrefix := hash[:2]
		layerPath := baseP + "/" + hashPrefix + "/" + hash

		statStr, err := config.Global.GlobalMnt.Statx(layerPath, cephfs.StatxSize, 0)
		if err != nil {
			panic(err)
		}
		
		
		intCastSize := (int64)(statStr.Size)

		
		layers = append(layers, veeOne.Descriptor{
			MediaType: veeOne.MediaTypeImageLayer,
			Digest:    openconcept.Digest(hash),
			Size:      intCastSize,
		})


		annotationMap["com.preallocate.me"] = hash
	}

	v := veeOne.Manifest{
		Versioned: specs.Versioned{SchemaVersion: 2},
		MediaType: veeOne.MediaTypeImageManifest, // application/vnd.oci.image.manifest.v1+json
		Config: veeOne.Descriptor{
			MediaType: veeOne.MediaTypeImageConfig,
			Digest:    openconcept.Digest("addSaneDefaults"),
			Size:      0,
		},
		Layers: layers,
		Annotations: annotationMap,
	}

	
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}