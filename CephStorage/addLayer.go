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
	oldMfest.Layers = append(oldMfest.Layers, ocispec.Descriptor{
		MediaType: "application/vnd.oci.image.layer.v1.tar",
		Digest: dgst,
		Size: size,
		Annotations: map[string]string{
			"containerd.io/snapshot/hash/": hash,
			"containerd.io/snapshot/version/": tagName,
			"containerd.io/uncompressed": hash,
		},
	})


	if oldMfest.Annotations == nil {
		oldMfest.Annotations = map[string]string{}
	}
	
	oldMfest.Annotations["containerd.io/snapshot.envVolPath"] = volPath

	return oldMfest,nil

}