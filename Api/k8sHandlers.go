package Api

import (
	"net/http"
	"log"
	"bytes"
	"encoding/json"

	s "compressor/CephStorage"
	c "compressor/Configuration"
)



type jsonReqSchema struct{
	EnvName string `json:"envName"`
	BaseImage string `json:"baseImage"`
}
	

//does env creation
func doCreateEnv(w http.ResponseWriter, r *http.Request) {
	log.Println("Creating env")
	//json parse
	defer r.Body.Close()

	var buf bytes.Buffer
	buf.ReadFrom(r.Body)
	var req jsonReqSchema

	json.Unmarshal(buf.Bytes(), &req)


	
	result, err := c.Global.InformConfig.NamespaceExists(req.EnvName)
	if err != nil {
		log.Println("ns check error failed with", req.EnvName, err)
		http.Error(w, "ns check failed", 500)
		return
	}

	if result {
		log.Println("ns exists")

		//success
		http.Error(w, "ns exists", 200)

	} else {
		log.Println("ns not exists")
		//creates ns

		err := c.Global.InformConfig.CreateNamespace(req.EnvName)
		if err != nil {
			http.Error(w, "create ns failed", 500)
			return
		}
		http.Error(w, "ns exists", 200)
	}

	log.Println("Created k8s resources, distribution backend management to be done here")


	globalConf := c.Global.GlobalMnt
	var mntWrap s.CephMount
	mntWrap.CephMount = &globalConf

	err = mntWrap.MakeRepo(req.EnvName, req.BaseImage)
	if err != nil {
		http.Error(w, "repo creation failed", 500)
		log.Printf("Function exited with error: %s\n", err.Error())
	}

	http.Error(w, "done", 200) //is OK, proceed
}