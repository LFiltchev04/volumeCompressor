package CephStorage

import (
	cnfg "compressor/Configuration"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"

	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)



func (cm *CephMount) AddDiffAiDi(oldMfest *ocispec.Manifest, diffId string, repoName string) (string, error){
	
	blobDatSpec := blobDataPathSpec{
		digest: oldMfest.Config.Digest,
	}

	imgConf := ocispec.Image{}
	
	rPath,err := PathFor(blobDatSpec)
	if err != nil {
		return "", err
	}

	//computes the new configuration
	fPath := path.Join(cnfg.Global.BaseConf.DistributionPV, rPath)
	file, err := cm.CephMount.Open(fPath, os.O_RDONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to open config file at %q: %w", fPath, err)
	}
	byes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read config file %q: %w", fPath, err)
	}
	
	if err := json.Unmarshal(byes, &imgConf); err != nil {
		return "", fmt.Errorf("failed to unmarshal config JSON: %w", err)
	}
	
	dgst := digest.FromString(diffId)
	imgConf.RootFS.DiffIDs = append(imgConf.RootFS.DiffIDs, dgst)

	jsonBytes, err := json.Marshal(imgConf)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(jsonBytes)
	hashStr := hex.EncodeToString(hash[:])
	println("hash for new config is: ", hashStr)
	dg, err := digest.Parse("sha256:" + hashStr)
	if err != nil {
		println("Failed to parse digest")
		return "", err
	}
	oldMfest.Config.Digest = dg

	println("old digest: ", oldMfest.Config.Digest.String())

	oldMfest.Config.Size = int64(len(jsonBytes))


	//writes the new blob to the image store

	newConfElem := blobDataPathSpec{
		digest: oldMfest.Config.Digest,
	}
	rPath, err = PathFor(newConfElem)
	if err != nil {
		return "", err
	}
	
	fPath = path.Join(cnfg.Global.BaseConf.DistributionPV, rPath)
	
	if err := cm.ensureDir(path.Dir(fPath)); err != nil {
		return "", fmt.Errorf("failed to create directory for %q: %w", fPath, err)
	}

	file, err = cm.CephMount.Open(fPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to create config file at %q: %w", fPath, err)
	}

	_, err = file.Write(jsonBytes)
	if err != nil {
		return "", fmt.Errorf("failed to write config data to %q: %w", fPath, err)
	}
	file.Close()
	
	
	lyrLink := layerLinkPathSpec{
		digest: oldMfest.Config.Digest,
	}
	rPath, err = PathFor(lyrLink)
	if err != nil {
		return "", err
	}

	fPath = path.Join(cnfg.Global.BaseConf.DistributionPV, rPath)
	cm.ensureDir(path.Dir(fPath))
	file, err = cm.CephMount.Open(fPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return "", err
	}
	_, err = file.Write([]byte(oldMfest.Config.Digest.String()))
	if err != nil {
		return "", fmt.Errorf("failed to write digest link at %q: %w", fPath, err)
	}
	file.Close()

	return hashStr, nil
}