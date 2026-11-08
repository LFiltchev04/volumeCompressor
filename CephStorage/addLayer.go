package CephStorage

import (
	//"os"
	//cnfg "compressor/Configuration"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	digest "github.com/opencontainers/go-digest"
)




//formats the new manifest to allow the snapshotter to lazy pull layers
func AddLayer(repoName string, tagName string, volPath string, hash string, size int64, oldMfest ocispec.Manifest) (ocispec.Manifest,error) {


	dgst, err := digest.Parse(hash)
	if err != nil {
		println("parse digest err: ", err.Error())
		return oldMfest, err
	}

	//get the last layer as is and check if it has unpack annotations 
	lastLayer := oldMfest.Layers[len(oldMfest.Layers)-1]
	_, exists:= lastLayer.Annotations["containerd.io/volPath"]
	if exists {
		println("last layer has volPath")
		lastLayer.Annotations = map[string]string{} //wipes them
	}


	oldMfest.Layers = append(oldMfest.Layers, ocispec.Descriptor{
		MediaType: "application/vnd.oci.image.layer.v1.tar",
		Digest: dgst,
		Size: size,
		Annotations: map[string]string{
			"containerd.io/snapshot/volPath/": volPath,
			"containerd.io/snapshot/hash/": hash,
			"containerd.io/snapshot/version/": tagName,
		},
	})




	return oldMfest,nil

}