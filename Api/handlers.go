package Api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gorilla/mux"

	"compressor/CephStorage"
	"compressor/Configuration"
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
	//gotta get rid of the hardlinks eventually
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

	hashPath := "/volumes/csi/csi-vol-3dbb4382-b70b-472d-9e7c-c31c9815841c/5030be57-07d6-4daa-ae31-13fae5206f5a/arch"
	cMnt.HardlinkBlob(hashPath,"rpo")

}


func doSnapshot(resp http.ResponseWriter, req *http.Request) {
	fmt.Println("doSnapshot: method=", req.Method, "Content-Length=", req.ContentLength)
	fmt.Println("doSnapshot: headers=", req.Header)

	bdyByts, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("doSnapshot: read body len=", len(bdyByts), "contents=", string(bdyByts))

	if len(bdyByts) == 0 {
		http.Error(resp, "empty request body", http.StatusBadRequest)
		return
	}

	type postBody struct {
		SnapshotID string `json:"SnapshotID"`
		VolumePath string `json:"VolumePath"`
		EnvName string `json:"EnvName"`
	}

	var postedData postBody
	if err := json.Unmarshal(bdyByts, &postedData); err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return
	}

	globalConf := Configuration.Global.GlobalMnt
	var mntWrap CephStorage.CephMount
	mntWrap.CephMount = &globalConf

	snapshotPath := postedData.VolumePath + "/.snap/" + postedData.EnvName;

	parrentDir := path.Dir(snapshotPath)
	strArr := strings.Split(parrentDir, "/")

	volumeName := strArr[len(strArr)-1]

	dirCont, err := mntWrap.CephMount.OpenDir(postedData.VolumePath + "/.snap/")
	if err != nil {
		fmt.Println("doSnapshot: error checking .snap directory:", err.Error())
		resp.WriteHeader(500)
		resp.Write([]byte("Error checking .snap directory: " + err.Error()))
		return
	}

	for {
		dentry, err := dirCont.ReadDir()
		if err != nil {
			errMsg := fmt.Sprintf("doSnapshot: error reading .snap directory: %v", err)
			fmt.Println(errMsg)
			resp.WriteHeader(500)
			resp.Write([]byte(errMsg))
			return
		}

		if dentry == nil {
			SnapshotCeph(postedData.EnvName, volumeName)
			break
		}
	
		if dentry.Name() == postedData.EnvName {
			break
		}
	
	}

	mntWrap.CompressSubvolume(snapshotPath, postedData.SnapshotID)


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
	fmt.Println("doDistribution: method=", req.Method, "Content-Length=", req.ContentLength)
	fmt.Println("doDistribution: headers=", req.Header)

	reqBody, err := io.ReadAll(req.Body)

	if err != nil {
		fmt.Println("Error on body read from doDistribution", err.Error())
		//revise theese error codes later
		resp.WriteHeader(500)
		return
	}

	fmt.Println("doDistribution: read body len=", len(reqBody), "contents=", string(reqBody))

	if len(reqBody) == 0 {
		http.Error(resp, "empty request body", http.StatusBadRequest)
		return
	}

	type reqJson struct {
		GenerationID string `json:"GenerationID"`
		VolumePath string `json:"VolumePath"`
		Username string `json:"Username"`
	}
	var jsonInstance reqJson

	if err := json.Unmarshal(reqBody, &jsonInstance); err != nil {
		fmt.Println("doDistribution: json unmarshal error:", err.Error())
		http.Error(resp, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

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
	cMnt.HardlinkBlob(rawBlobPath,jsonInstance.Username)

}



func doMakeEnv(resp http.ResponseWriter, req *http.Request){


	type reqJson struct {
		VolumePath string `json:"VolumePath"`
		Username string `json:"Username"`
		EnvName string `json:"EnvName"`
	}
	var jsonInstance reqJson

	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Println("Error on body read from doMakeEnv", err.Error())
		resp.WriteHeader(500)
		return
	}

	if err := json.Unmarshal(reqBody, &jsonInstance); err != nil {
		fmt.Println("doMakeEnv: json unmarshal error:", err.Error())
		http.Error(resp, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}


	
	globalConf := Configuration.Global.GlobalMnt
	var mntWrap CephStorage.CephMount
	mntWrap.CephMount = &globalConf

	err = mntWrap.CephMount.MakeDir(jsonInstance.VolumePath+ "/" + jsonInstance.EnvName, 0644)
	if err != nil {
		fmt.Println("doMakeEnv: error creating directory:", err.Error())
		resp.WriteHeader(500)
		resp.Write([]byte("Error creating directory: " + err.Error()))
		return
	}
	
	//simple symlink to get the other dirs to show up in alt repos

	resp.WriteHeader(200)
	resp.Write([]byte("Environment created successfully"))
}