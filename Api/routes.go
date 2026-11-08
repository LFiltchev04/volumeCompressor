package Api

import (
	"log"
	"net/http"
)

func Listener() {
	mux := http.NewServeMux()
	mux.HandleFunc("/test", doTest)
	mux.HandleFunc("POST /snapshotState", doSnapshot)
	mux.HandleFunc("POST /repoPush", doDistribution)

	//the K8s stuff is here
	mux.HandleFunc("POST /createEnv", doCreateEnv)
	//mux.HandleFunc("POST /deleteEnv", doDeleteEnv)

	mux.HandleFunc("GET /repoIndexBlobs", doGetIndexList)
	//mux.HandleFunc("POST /advanceTag", doTagAdv)

	//mux.HandleFunc("/healthz/{service:[a-zA-Z0-9.-]+}", doHealthz)
	log.Println("Starting server on :3001")
	if err := http.ListenAndServe(":3001", mux); err != nil {
		log.Fatal("HTTP server failed: ", err)
	}
}
