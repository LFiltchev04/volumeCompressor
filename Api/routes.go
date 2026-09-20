package Api

import "net/http"



func Listener() {
	mux := http.NewServeMux()
	mux.HandleFunc("/test", doTest)
	mux.HandleFunc("/snapshotState", doSnapshot)
	mux.HandleFunc("POST /repoPush", doDistribution)
	mux.HandleFunc("POST /makeSubenv", doMakeEnv)
	
	//mux.HandleFunc("/healthz/{service:[a-zA-Z0-9.-]+}", doHealthz)
	http.ListenAndServe(":3000", mux)
	_ = mux
}
