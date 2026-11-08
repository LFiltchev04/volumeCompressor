package CephStorage

import (
	"os"
	"fmt"
	config "compressor/Configuration"
	digest "github.com/opencontainers/go-digest"
)





func (cm *CephMount) WriteNewTag(newTagName string, repo string, hash string) error {

	basePV := config.Global.BaseConf.DistributionPV
	dgst, err := digest.Parse("sha256:"+hash)
	if err != nil {
		return err
	}

	elem := manifestTagIndexEntryPathSpec{
		name: repo,
		tag: newTagName,
		revision: dgst,
	}
	rPath, err := PathFor(elem)
	
	if err != nil {
		fmt.Println("error getting path")
		return err
	}
	
	fPath := basePV + rPath

	err = cm.ensureDir(fPath)
	if err != nil {
		if os.IsExist(err) {
			fmt.Println("dir does not exist")
		}else{
			fmt.Println("error ensuring dir")
			return err
		}
		
	}
	fmt.Println("ensured dir")

	file, err := cm.CephMount.Open(fPath+ "/link", os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("error opening file with path: %s", fPath)
		return err
	}
	defer file.Close()

	_, err = file.Write([]byte("sha256:"+hash))
	if err != nil {
		return err
	}

	crrElem := manifestTagPathSpec{
		name: repo,
		tag: newTagName,
	}

	rPath, err = PathFor(crrElem)
	if err != nil {
		return err
	}
	fPath = basePV + rPath + "/current"
	err = cm.ensureDir(fPath)
	if err != nil {
		return err
	}

	file, err = cm.CephMount.Open(fPath+"/link", os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	_, err = file.Write([]byte("sha256:"+hash))
	if err != nil {
		fmt.Printf("Errors on the current subdir write path: %s\n", fPath+"/link")
		return err
	}

	//now writing to revisions
	revElem := manifestRevisionPathSpec{
		name: repo,
		revision: dgst,
	}
	rPath, err = PathFor(revElem)
	if err != nil {
		return err
	}
	fPath = basePV + rPath
	err = cm.ensureDir(fPath)
	if err != nil {
		fmt.Println("error ensuring dir for revision")
		return err
	}

	file, err = cm.CephMount.Open(fPath+ "/link", os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("error opening revision link")
		return err
	}

	_, err = file.Write([]byte("sha256:"+hash))
	if err != nil {
		return err
	}
	println("wrote to ")
	
	return nil
}