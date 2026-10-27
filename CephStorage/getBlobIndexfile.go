package CephStorage


import(
	cnf "compressor/Configuration"
	"fmt"
)






func (cm *CephMount) GetBlobIndex(repoName string, blobHash string) (string, error) {



	repoPathSpc := manifestRevisionsPathSpec{
		name: repoName,
	}

	repoPath, err := PathFor(repoPathSpc)
	if err != nil {
		return "", err
	}
	fPath := cnf.Global.BaseConf.DistributionPV + repoPath


	//has to have just one entry
	dir, err := cm.CephMount.OpenDir(fPath)
	if err != nil {
		return "", err
	}


	var hash string

	//has to have only one dir entry, a hash leading to the linkfile containing file and the .. dir entry
	for n := 0; n < 4; n++ {
		
		dentries, err := dir.ReadDir()
		if dentries == nil {
			break
		}
		if err != nil {
			return "", err
		}
		
		if dentries.Name() == ".." {
			continue
		}

		if dentries.Name() == "."{
			continue
		}


		hash = dentries.Name()
	}

	linkfileContaining := fPath + "/" + hash
	defer dir.Close()

	contDir, err := cm.CephMount.OpenDir(linkfileContaining)
	if err != nil {
		return "", err
	}
	defer contDir.Close()

	for dentry, err := contDir.ReadDir(); dentry != nil; dentry, err = contDir.ReadDir() {
		if err != nil {
			return "", err
		}

		if dentry.Name() == blobHash {
			return dentry.Name(), nil
		}


	}

	return "", fmt.Errorf("not found")
}