package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	//"github.com/aws/aws-sdk-go/aws"
	//"github.com/aws/aws-sdk-go/aws/session"
	//"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gorilla/mux"
	"github.com/jamespearly/loggly"
	"github.com/joho/godotenv"
)

type information struct {
	Count        int64 `json:"Count"`
	Items        Item  `json:"Items"`
	ScannedCount int64 `json:"ScannedCount"`
}

type Item struct {
	Timestamp         int64  `json:"timestamp"`
	Id                string `json:"id"`
	Rank              string `json:"rank"`
	Symbol            string `json:"symbol"`
	Name              string `json:"name"`
	Supply            string `json:"supply"`
	MaxSupply         string `json:"maxSupply"`
	MarketCapUsd      string `json:"marketCapUsd"`
	VolumeUsd24Hr     string `json:"volumeUsd24Hr"`
	PriceUsd          string `json:"priceUsd"`
	ChangePercent24hr string `json:"changePercent24hr"`
	Vwap24Hr          string `json:"vwap24Hr"`
	Explorer          string `json:"explorer"`
}

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

func all(w http.ResponseWriter, r *http.Request) {

	err := godotenv.Load("app.env")
	if err != nil {
		fmt.Println("Error loading the .env file")
	}

	//define a client for loggly
	client := loggly.New("nazartrut")

	//see if the path is correct, if not throw a 400 bad request error
	fmt.Println(r.URL.Path)
	if r.URL.Path != "/ntrut/all" {
		client.Send("error", "Error, bad request for the endpoint /ntrut/all")
		//send back 400 status code
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("400 - Bad Request")
	}

	// Create an AWS session for US East 1. REE
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}))

	// Create a DynamoDB instance
	db := dynamodb.New(sess)
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		client.Send("error", "Incorrect Method, Needs to be GET. Endpoint: /ntrut/all")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode("405 - Method Not Allowed")
	} else {
		//query the db
		params := &dynamodb.ScanInput{
			TableName: aws.String("Crypto"),
		}

		//scan the db, this returns all items
		result, err := db.Scan(params)
		if err != nil {
			client.Send("error", "Error getting all items from Crpto Table. Endpoint: /ntrut/all")
		} else {
			//create an array, then unshall the gotten items from the scan
			var info []Item
			err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &info)
			if err != nil {
				_ = client.Send("error", "Got error unmarshalling items. Endpoint: /ntrut/all")
			}

			//success
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(info)
			client.Send("info", "Successfully returned all items from Cryto dynamoDB table. Length of data: "+strconv.Itoa(len(info)))
		}
	}
}

func main() {
	//define a client for logglyy
	r := mux.NewRouter()
	r.HandleFunc("/ntrut/status", statusHandler).Methods("GET")
	r.HandleFunc("/ntrut/all", all).Methods("GET")
	http.ListenAndServe(":8080", r)

}
