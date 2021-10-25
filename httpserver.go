package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/gorilla/mux"
	"github.com/jamespearly/loggly"
	"github.com/joho/godotenv"
	"net/http"
	"regexp"
	"strconv"
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
	//define a client for loggly
	client := loggly.New("nazartrut")

	//see if the path is correct, if not throw a 400 bad request error
	if r.URL.Path != "/ntrut/status" {
		client.Send("error", "Error, bad request for the endpoint /ntrut/status")
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
		client.Send("error", "Incorrect Method, Needs to be GET. Endpoint: /ntrut/status")
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
			client.Send("error", "Error getting all items from Crpto Table. Endpoint: /ntrut/status")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode("500 - Internal Server Error")
		} else {
			data := information{
				Table:       "Crypto",
				RecordCount: result.Count,
			}

			//marshall map the struct
			av, err := dynamodbattribute.MarshalMap(data)
			if err != nil {
				client.Send("error", "Got error marshalling new data item. Endpoint: /ntrut/status")
			}

			//success
			fmt.Println(av)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data)
			client.Send("info", "Successfully returned all items from Cryto dynamoDB table. Length of data: "+strconv.Itoa(len(av)))
		}
	}
}

func all(w http.ResponseWriter, r *http.Request) {
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

/*search by symbol*/
func search(w http.ResponseWriter, r *http.Request) {
	//define a client for loggly
	client := loggly.New("nazartrut")
	//see if the path is correct, if not throw a 400 bad request error
	fmt.Println(r.URL.Path)
	if r.URL.Path != "/ntrut/search" {
		client.Send("error", "Error, bad request for the endpoint /ntrut/search")
		//send back 400 status code
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("400 - Bad Request")
		return
	}

	keys, err := r.URL.Query()["key"]
	fmt.Println(keys)
	if !err || len(keys[0]) < 1 {
		client.Send("error", "the query parameter is empty. Endpoint: /ntrut/search")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("400 - Bad Request")
		return
	}
	//query returns array but we only want the single item
	key := keys[0]

	//if match is true, then it is not fully a string and throw error and return 400
	match, _ := regexp.MatchString("\\D+", key)
	if match == true {
		client.Send("error", "the query parameter can only have numbers. Endpoint: /ntrut/search")
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
		client.Send("error", "Incorrect Method, Needs to be GET. Endpoint: /ntrut/search")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode("405 - Method Not Allowed")
	} else {
		//query the db
		result, err := db.GetItem(&dynamodb.GetItemInput{
			TableName: aws.String("Crypto"),
			Key: map[string]*dynamodb.AttributeValue{
				"timestamp": {
					N: aws.String(key),
				},
			},
		})

		if err != nil {
			client.Send("error", "Error getting all items from Crpto Table. Endpoint: /ntrut/search")
		} else {

			//marshall map the struct
			info := Item{}

			// Unmarshal all resulting scan data into Information struct.
			err := dynamodbattribute.UnmarshalMap(result.Item, &info)
			if err != nil {
				_ = client.Send("error", "Got error marshalling new movie item")
			}

			if info.Id == "" {
				client.Send("error", "Item not found in database Endpoint: /ntrut/search/")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode("404 - Item Not Found")
			} else {
				//success
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(info)
				client.Send("info", "Successfully returned the searched item from Cryto dynamoDB table. Length of data: "+strconv.Itoa(len(result.Item)))
			}
		}
	}
}

func main() {
	//define a client for logglyy
	err := godotenv.Load("./app.env")
	if err != nil {
		fmt.Println("Error loading the .env file")
		return
	}

	r := mux.NewRouter()
	r.HandleFunc("/ntrut/status", statusHandler).Methods("GET")
	r.HandleFunc("/ntrut/all", all).Methods("GET")
	r.HandleFunc("/ntrut/search", search).Methods("GET")
	http.ListenAndServe(":8080", r)

}
