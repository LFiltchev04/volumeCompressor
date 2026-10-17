package Api

import "net/http"



func Listener() {
	mux := http.NewServeMux()
	mux.HandleFunc("/test", doTest)
	mux.HandleFunc("POST /snapshotState", doSnapshot)
	mux.HandleFunc("POST /repoPush", doDistribution)

	//mux.HandleFunc("POST /advanceTag", doTagAdv)
	//mux.HandleFunc("POST /makeSubenv", doMakeEnv)
	
	//mux.HandleFunc("/healthz/{service:[a-zA-Z0-9.-]+}", doHealthz)
	http.ListenAndServe(":3001", mux)
}
