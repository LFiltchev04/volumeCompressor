package Api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/ceph/go-ceph/cephfs"

	"github.com/gorilla/mux"

	"compressor/CephStorage"
	"compressor/Configuration"
	//help "compressor/Helpers"
)

func doTest(resp http.ResponseWriter, req *http.Request){
	//bdy, _ := io.ReadAll(req.Body)
	
	//type postBody struct {
	//	SnapshotID string `json:"SnapshotID"`
	//	VolumePath string `json:"VolumePath"`
	//}

	//var elem postBody
	//json.Unmarshal(bdy, &elem)

	mnt := Configuration.Global.GlobalMnt
	dirEnt, _ := mnt.OpenDir("/volumes/csi/csi-vol-3dbb4382-b70b-472d-9e7c-c31c9815841c/5030be57-07d6-4daa-ae31-13fae5206f5a/arch")
	if dirEnt == nil {
		println("this was nil?")
	}
	var archFile string
	for {
		dentry, err := dirEnt.ReadDir()
		if err != nil {
			println("error in test dir iter: ", err.Error())
			break
		}
		if dentry == nil {
			break
		}
		

		archFile = dentry.Name()
	}

	println("derived arch file: ", archFile)

	//var cMnt CephStorage.CephMount
	//cMnt.CephMount = &Configuration.Global.GlobalMnt

	//hashPath := "/volumes/csi/csi-vol-3dbb4382-b70b-472d-9e7c-c31c9815841c/5030be57-07d6-4daa-ae31-13fae5206f5a/arch"
	//cMnt.HardlinkBlob(hashPath,"rpo")
//a
}







func doSnapshot(resp http.ResponseWriter, req *http.Request) {
	bdyByts, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Method != http.MethodPost {
		resp.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	
	///IMPORTANT!!!!!!
	type postBody struct {
		RepoName string `json:"repoName"`
		Version string `json:"version"`
		VolumePath string `json:"volumePath"`
	}

	var postedData postBody
	if err := json.Unmarshal(bdyByts, &postedData); err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return
	}

	globalConf := Configuration.Global.GlobalMnt
	var mntWrap CephStorage.CephMount
	mntWrap.CephMount = &globalConf

	mntWrap.CompressSubvolume(postedData.VolumePath, postedData.Version)
	VolumePath := postedData.VolumePath	
	
	//hardlinks it from the PV into the redable directory
	hash, err := mntWrap.HardlinkBlob(VolumePath + "/arch/" + postedData.Version + ".tar", postedData.RepoName)
	if err != nil {
		println("hardlink err: ", err.Error())
	}
	println("hash is: ", hash)

	//adds a new manifest tapped to the repo name

	mFestOld, err := mntWrap.LoadManifest(postedData.RepoName)
	if err != nil {
		println("load manifest err: ", err.Error())
		return
	}

	//gotta pull the size out too
	mFestNew, err := CephStorage.AddLayer(postedData.RepoName,postedData.Version, hash, -10, *mFestOld)
	if err != nil {
		println("add layer err: ", err.Error())
	}

	err = mntWrap.WriteToLayers(postedData.RepoName, postedData.Version, hash)
	if err != nil {
		println("write to layers err: ", err.Error())
	}

	byteCast, err := json.Marshal(mFestNew)
	if err != nil {
		println("marshal err: ", err.Error())
	}

	hash, err = mntWrap.WriteBlob(&byteCast, postedData.RepoName)
	if err != nil {
		println("write blob err: ", err.Error())
	}
	

	resp.WriteHeader(http.StatusOK)
	_, _ = resp.Write([]byte("ok"))
}






func doHealthz(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	srvcType := vars["service"]

	if srvcType == "ceph" {
		urlFull := Configuration.Global.BaseConf.MonIPs[0] + "/api/health/minimal"

		healthResp, err := http.Get(urlFull)
		if err != nil {
			println("Health check failed with: ", err.Error())
		} else if healthResp.StatusCode == http.StatusOK {
			// passed ok; nothing else needed here
		}
		if healthResp != nil && healthResp.Body != nil {
			_ = healthResp.Body.Close()
		}
	}
	_, _ = resp.Write([]byte("ok"))
}


func doDistribution(resp http.ResponseWriter, req *http.Request){
	
	reqBody, err := io.ReadAll(req.Body)
	
	if err != nil {
		println("Error on body read from doDistribution", err.Error())
		//revise theese error codes later
		resp.WriteHeader(500)
		return
	}

	type reqJson struct {
		GenerationID string `json:"GenerationID"`
		VolumePath string `json:"VolumePath"`
		Username string `json:"Username"`
	}
	var jsonInstance reqJson

	json.Unmarshal(reqBody,&jsonInstance)

	println("vol pth: ", jsonInstance.VolumePath)

	rawBlobPath := jsonInstance.VolumePath + "/" + jsonInstance.GenerationID
	println("raw blob path: ", rawBlobPath)
	rawBlobFd, err :=Configuration.Global.GlobalMnt.Open(rawBlobPath,os.O_RDWR,0644)
	
	if err != nil { 
		println("Raw blob reader is throwing an error: ", err.Error() + " on path: ", rawBlobPath)
		resp.WriteHeader(500)

		resp.Write([]byte("Raw blob reader is throwing error: "+ err.Error()))
		return
	}

	if rawBlobFd == nil {
		println("Cant open raw blob path, is nil")
		resp.WriteHeader(500)
		resp.Write([]byte("Cant open raw blob path, is nil"))
		
		return
	}

	var cMnt CephStorage.CephMount
	cMnt.CephMount = &Configuration.Global.GlobalMnt
	cMnt.HardlinkBlob(rawBlobPath, jsonInstance.VolumePath)

}


/*
func doManifestOnly(resp http.ResponseWriter, req *http.Request){
	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		println("Error on body read from doManifestOnly", err.Error())
		resp.WriteHeader(500)
		return
	}

	globalConf := Configuration.Global.GlobalMnt
	var mntWrap CephStorage.CephMount
	mntWrap.CephMount = &globalConf

	type reqJson struct {
		TagName string `json:"GenerationID"`
		RepoID   string `json:"VolumePath"`
		BlobList []string `json:"BlobList"`
	}
	var jsonInstance reqJson

	json.Unmarshal(reqBody, &jsonInstance)

	var uint64Sizes []int64
	for _, blb := range jsonInstance.BlobList {
		
		size, err := mntWrap.GetSize(blb)

		if err != nil {
			println("Error getting size for blob: ", blb, " error: ", err.Error())
			resp.WriteHeader(500)
			return
		}

		//its big enough
		intSize := int64(size)
		uint64Sizes = append(uint64Sizes, intSize)
	}
															//hope thats right
	mFest, err:= help.FormatManifest(jsonInstance.BlobList, uint64Sizes, "application.vnd.oci.image.manifest.v1+json")
	if err != nil {
		println("Error formatting manifest: ", err.Error())
		resp.WriteHeader(500)
		return
	}
	byteArr, _ := json.Marshal(mFest)
	
	if err := mntWrap.WriteBlob(&byteArr); err != nil {
		println("Error writing manifest blob: ", err.Error())
		resp.WriteHeader(500)
		return
	}

	resp.WriteHeader(200)
}

*/

type indexeReq struct {
	RepoName string `json:"repoName"`
	TagVersion string `json:"tagVersion"`

}
//fetches the index blobs for a repo
func doGetIndexList(resp http.ResponseWriter, req *http.Request){
	index := indexeReq{}
	json.NewDecoder(req.Body).Decode(&index)
	println("repo: ", index.RepoName)
	

}


type indexBlobReq struct {
	RepoName string `json:"repoName"`
	BlobID string `json:"blobID"`
}
func doGetIndexBlob(resp http.ResponseWriter, req *http.Request){
	reqBody, _ := io.ReadAll(req.Body)
	var jsonInstance indexBlobReq
	json.Unmarshal(reqBody, &jsonInstance)
	println("blob: ", jsonInstance.BlobID)


	globalConf := Configuration.Global.GlobalMnt
	var mntWrap CephStorage.CephMount
	mntWrap.CephMount = &globalConf

	symlinkPath, err := mntWrap.GetBlobIndex(jsonInstance.RepoName, jsonInstance.BlobID)
	if err != nil {
		println("Error: ", err.Error())
		resp.WriteHeader(500)
		return
	}
	println("symlink: ", symlinkPath)


	reader, err := mntWrap.CephRead(symlinkPath)
	if err != nil {
		resp.WriteHeader(500)
		return
	}
	defer reader.Close()

	// Statx follows the symlink (no AtSymlinkNofollow), so this is the real file's size
	stat, err := mntWrap.CephMount.Statx(symlinkPath, cephfs.StatxSize, 0)
	if err != nil {
		resp.WriteHeader(500)
		return
	}
	endfileSize := int64(stat.Size)

	resp.Header().Set("Content-Type", "application/octet-stream")
	resp.Header().Set("Content-Length", strconv.FormatInt(endfileSize, 10))
	resp.WriteHeader(200)
	io.Copy(resp, reader)

}