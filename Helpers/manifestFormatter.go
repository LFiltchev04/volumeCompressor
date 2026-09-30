package Helpers


import(
	"github.com/opencontainers/image-spec/specs-go/v1"
    digest "github.com/opencontainers/go-digest"

)


//the arrays must be aligned or it will break, stuff for upstream
func FormatManifest(blobList []string, szArr []int64 , mediaType string) (*v1.Manifest, error) {
	//will have a static config for now
	manifest := &v1.Manifest{
		Config: v1.Descriptor{},
		MediaType: mediaType,
		Layers:    []v1.Descriptor{},
	}


	for ix, blob := range blobList {
		
		hashWrp, _ := digest.Parse(blob)
		
		manifest.Layers = append(manifest.Layers, v1.Descriptor{
			MediaType: mediaType,
			Digest:    hashWrp,
			Size:      szArr[ix],
		})
	}


	return manifest, nil
}



