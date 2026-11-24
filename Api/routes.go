package Api

import (
	"log"
	"net/http"
)

func Listener() {
	mux := http.NewServeMux()
	//normal fs manipulation section routes
	mux.HandleFunc("/test", doTest)
	mux.HandleFunc("POST /snapshotState", doSnapshot) //makes a new tag of an already existing environment
	mux.HandleFunc("POST /repoPush", doDistribution)

	//the K8s stuff is here
	mux.HandleFunc("POST /createEnv", doCreateEnv) //makes a new environment from an existing base
	//mux.HandleFunc("POST /deleteEnv", doDeleteEnv)

	mux.HandleFunc("GET /repoIndexBlobs", doGetIndexList)
	//mux.HandleFunc("POST /advanceTag", doTagAdv)

	//mux.HandleFunc("/healthz/{service:[a-zA-Z0-9.-]+}", doHealthz)
	log.Println("Starting server on :3001")
	if err := http.ListenAndServe(":3001", mux); err != nil {
		log.Fatal("HTTP server failed: ", err)
	}
}
