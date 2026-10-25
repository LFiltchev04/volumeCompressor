package Api

import (
	"net/http"
	"log"
	"bytes"
	"encoding/json"

	k "compressor/kubWatcher"
	s "compressor/CephStorage"
)



type jsonReqSchema struct{
	EnvName string `json:"envName"`
	BaseImage string `json:"baseImage"`
}
	
var ic = &k.InformerConfig{}

//does env creation
func doCreateEnv(w http.ResponseWriter, r *http.Request) {
	log.Println("Creating env")
	//json parse
	defer r.Body.Close()

	var buf bytes.Buffer
	buf.ReadFrom(r.Body)
	var req jsonReqSchema

	json.Unmarshal(buf.Bytes(), &req)


	result, err := ic.NamespaceExists(req.EnvName)
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

		err := ic.CreateNamespace(req.EnvName)
		if err != nil {
			http.Error(w, "create ns failed", 500)
			return
		}
		http.Error(w, "ns exists", 200)
	}

	log.Println("Created k8s resources, distribution backend management to be done here")







	http.Error(w, "done", 200) //is OK, proceed
}