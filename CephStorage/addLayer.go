package CephStorage

import (
	//"os"
	//cnfg "compressor/Configuration"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	digest "github.com/opencontainers/go-digest"
)




//formats the new manifest to allow the snapshotter to lazy pull layers
func AddLayer(repoName string, tagName string, volPath string, hash string, size int64, oldMfest ocispec.Manifest) (ocispec.Manifest,error) {



	oldMfest.Layers = append(oldMfest.Layers, ocispec.Descriptor{
		Digest: digest.FromString(hash),
		Size: size,
		Annotations: map[string]string{
			"io.containerd/snapshot/hash/": hash,
			"io.containerd/snapshot/version/": tagName,
		},
	})


	if oldMfest.Annotations == nil {
		oldMfest.Annotations = map[string]string{}
	}
	
	oldMfest.Annotations["io.containerd/snapshot/envVolPath/"] = volPath

	return oldMfest,nil

}