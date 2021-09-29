package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/jamespearly/loggly"
)

func statusHandler(w http.ResponseWriter, r *http.Request) {

	//loggly stuff
	err := os.Setenv("LOGGLY_TOKEN", "f6522fc1-420b-4614-8cfa-013b881bac56")
	if err != nil {
		return
	}
	client := loggly.New("nazartrut")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	//json that will be returned in the get request
	outputJson := `{"status":  0, "system-time": 0}`
	out := map[string]interface{}{}
	json.Unmarshal([]byte(outputJson), &out)

	out["status"] = http.StatusOK
	out["system-time"] = time.Now()

	output, err := json.Marshal(out)
	if err != nil {
		client.Send("error", "Error marshalling the output Json")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	//json that is send back to loggly
	logglyJson := `{"status":  0, "method-type": "", "source": "", "request-path": ""}`
	out2 := map[string]interface{}{}
	json.Unmarshal([]byte(logglyJson), &out2)

	//set the logglyjson
	out2["status"] = http.StatusOK
	out2["method-type"] = "GET"
	out2["source"] = r.RemoteAddr
	out2["request-path"] = r.URL.Path

	logglyOutput, err := json.Marshal(out2)
	if err != nil {
		client.Send("error", "Error marshalling the loggly output Json")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	client.Send("info", string(logglyOutput))
	io.WriteString(w, string(output))
}

func main() {
	//define a client for loggly
	r := mux.NewRouter()
	r.HandleFunc("/ntrut/status", statusHandler).Methods("GET")
	http.ListenAndServe(":8080", r)
}
