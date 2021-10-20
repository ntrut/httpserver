package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"net/http"
	"strconv"
	//"github.com/aws/aws-sdk-go/aws"
	//"github.com/aws/aws-sdk-go/aws/session"
	//"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gorilla/mux"
	"github.com/jamespearly/loggly"
	"github.com/joho/godotenv"
)

type information struct {
	Table       string `json:"table"`
	RecordCount *int64 `json:"recordCount"`
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

	err := godotenv.Load("app.env")
	if err != nil {
		fmt.Println("Error loading the .env file")
	}

	//define a client for loggly
	client := loggly.New("nazartrut")

	//see if the path is correct, if not throw a 400 bad request error
	fmt.Println(r.URL.Path)
	if r.URL.Path != "/805857442/status" {
		client.Send("error", "Error, bad request for the endpoint /805857442/status")
		//send back 400 status code
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("400 - Bad Request")
		return
	}

	// Create an AWS session for US East 1. REE
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}))

	// Create a DynamoDB instance
	db := dynamodb.New(sess)
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		client.Send("error", "Incorrect Method, Needs to be GET. Endpoint: /805857442/status")
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
			client.Send("error", "Error getting all items from Crpto Table. Endpoint: /805857442/status")
		} else {
			data := information{
				Table:       "Crypto",
				RecordCount: result.Count,
			}

			//marshall map the struct
			av, err := dynamodbattribute.MarshalMap(data)
			if err != nil {
				client.Send("error", "Got error marshalling new data item. Endpoint: /805857442/status")
			}

			//success
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data)
			client.Send("info", "Successfully returned all items from Cryto dynamoDB table. Length of data: "+strconv.Itoa(len(av)))
		}
	}
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
	if r.URL.Path != "/805857442/all" {
		client.Send("error", "Error, bad request for the endpoint /805857442/all")
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
		client.Send("error", "Incorrect Method, Needs to be GET. Endpoint: /805857442/all")
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
			client.Send("error", "Error getting all items from Crpto Table. Endpoint: /805857442/all")
		} else {
			//create an array, then unshall the gotten items from the scan
			var info []Item
			err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &info)
			if err != nil {
				_ = client.Send("error", "Got error unmarshalling items. Endpoint: /805857442/all")
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
	r.HandleFunc("/805857442/status", statusHandler).Methods("GET")
	r.HandleFunc("/805857442/all", all).Methods("GET")
	http.ListenAndServe(":8080", r)

}
