package CephStorage

import (
	//"os"
	//cnfg "compressor/Configuration"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	digest "github.com/opencontainers/go-digest"
)




//formats the new manifest to allow complete 
func AddLayer(repoName string, tagName string, hash string, size int64, oldMfest ocispec.Manifest) (ocispec.Manifest,error) {



	oldMfest.Layers = append(oldMfest.Layers, ocispec.Descriptor{
		Digest: digest.FromString(hash),
		Size: size,
		Annotations: map[string]string{
			"com.LazyPull.blobHash": hash,
		},
	})




	return oldMfest,nil

}