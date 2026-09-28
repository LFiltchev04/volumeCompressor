package Api

import (
	"compressor/CephStorage"
	"compressor/Configuration"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"
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

	var cMnt CephStorage.CephMount
	cMnt.CephMount = &Configuration.Global.GlobalMnt

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
		DumpName string `json:"DumpName"` 
		UserID string `json:"UserID"`
		VolumePath string `json:"VolumePath"`
	}

	var postedData postBody
	if err := json.Unmarshal(bdyByts, &postedData); err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return
	}

	globalConf := Configuration.Global.GlobalMnt
	var mntWrap CephStorage.CephMount
	mntWrap.CephMount = &globalConf
	mntWrap.CompressSubvolume(postedData.VolumePath, postedData.UserID, postedData.DumpName)


	//for now i am assuming that the path is right, will see
	



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
	rawBlobFd, err :=Configuration.Global.GlobalMnt.Open(rawBlobPath,os.O_RDONLY,0644)
	
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
	cMnt.HardlinkBlob(rawBlobPath, rawBlobPath, jsonInstance.Username)

}