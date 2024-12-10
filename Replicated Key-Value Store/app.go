package main

// imported libraries
import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Global variables
var kvs = make(map[string]interface{})
var stringToIntMap = make(map[string]int)
var ipAddr = os.Getenv("SOCKET_ADDRESS")

// VIEW
var (
	view      = make(map[string]bool) // Tracks the view of replicas (socket-address -> bool)
	viewMutex = sync.Mutex{}          // Mutex to ensure thread-safe access to the view map
)

// Initialize the view based on the VIEW environment variable
func initializeView() {
	viewMutex.Lock()
	defer viewMutex.Unlock()

	//Load the initial list of replicas from the VIEW environment variable
	initialReplicas := strings.Split(os.Getenv("VIEW"), ",")
	for _, replica := range initialReplicas {
		if replica != "" {
			view[replica] = true
		}
	}

	// Fetch the latest view from any active replica
	for replica := range view {
		if replica != ipAddr {
			resp, err := http.Get("http://" + replica + "/view")
			if err != nil {
				//Error handling
				fmt.Printf("Failed to sync view from %s: %v\n", replica, err)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				var response map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&response)
				updatedView, ok := response["view"].([]interface{})
				if ok {
					view = make(map[string]bool)
					for _, r := range updatedView {
						view[r.(string)] = true
					}
				}
				break
			}
		}
	}

	// Fetch all key-value pairs from an active replica
	for replica := range view {
		if replica != ipAddr {
			KVSSyncing(replica)
			break
		}
	}

	// Notify other replicas of this replica's presence
	for replica := range view {
		if replica != ipAddr {
			go func(replica string) {
				requestBody, _ := json.Marshal(map[string]string{
					"socket-address": ipAddr,
				})
				req, err := http.NewRequest("PUT", "http://"+replica+"/view", strings.NewReader(string(requestBody)))
				if err != nil {
					//Error handling
					fmt.Printf("Failed to create PUT /view request for %s: %v\n", replica, err)
					return
				}
				req.Header.Set("Content-Type", "application/json")

				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					//Error handling
					fmt.Printf("Failed to send PUT /view to %s: %v\n", replica, err)
					return
				}
				defer resp.Body.Close()
				fmt.Printf("Rejoin broadcast sent to %s: %d\n", replica, resp.StatusCode)
			}(replica)
		}
	}
}

// Function to synchronize KVS and its metadata
func KVSSyncing(replica string) {
	fmt.Printf("Fetching key-value store and metadata from %s\n", replica)
	//Send an HTTP GET request to the /kvs/sync endpoint of the specified replica
	resp, err := http.Get("http://" + replica + "/kvs/sync")
	if err != nil {
		//Error Handling
		fmt.Printf("Failed to fetch key-value store from %s: %v\n", replica, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var response map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			//Error Handling
			fmt.Printf("Failed to decode key-value store response: %v\n", err)
			return
		}

		kvPairs, ok := response["kvs"].(map[string]interface{})
		if ok {
			for key, value := range kvPairs {
				kvs[key] = value
			}
		}

		metadata, ok := response["causal-metadata"].(map[string]interface{})
		if ok {
			for key, value := range metadata {
				stringToIntMap[key] = int(value.(float64)) // Ensure proper type conversion
			}
		}

		fmt.Printf("Synchronization completed: keys=%v, metadata=%v\n", kvs, stringToIntMap)
	} else {
		//Error Handling
		fmt.Printf("Failed to fetch key-value store from %s: status %d\n", replica, resp.StatusCode)
	}
}

// Function to sync KVS with its metadata
func syncHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"kvs":             kvs,
		"causal-metadata": stringToIntMap,
	}
	fmt.Printf("Responding with key-value store and metadata: %v\n", response)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Updating View
func updateViewHandler(w http.ResponseWriter, r *http.Request) {
	var body map[string][]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		//Error Handling
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updatedView, exists := body["view"]
	if !exists {
		//Error Handling
		http.Error(w, "Missing view in request body", http.StatusBadRequest)
		return
	}

	//Update the view map to match the updated view
	viewMutex.Lock()
	view = make(map[string]bool) // Reset the view
	for _, replica := range updatedView {
		view[replica] = true
	}
	viewMutex.Unlock()

	//Success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"result": "view updated"})
}

// Broadcasting updated view to other replicas
func broadcastViewUpdate() {
	viewMutex.Lock()
	currentView := make([]string, 0, len(view))
	for replica := range view {
		currentView = append(currentView, replica)
	}
	viewMutex.Unlock()

	//Iteration
	for addr := range view {
		go func(addr string) {
			requestBody, _ := json.Marshal(map[string]interface{}{
				"view": currentView,
			})
			req, err := http.NewRequest("PUT", "http://"+addr+"/view/update", strings.NewReader(string(requestBody)))
			if err != nil {
				//Error Handling
				fmt.Printf("Failed to create request for %s: %v\n", addr, err)
				return
			}
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil || resp.StatusCode != http.StatusOK {
				//Error Handling
				fmt.Printf("Failed to send view update to %s: %v\n", addr, err)
				return
			}
			//Success
			defer resp.Body.Close()
			fmt.Printf("View update sent to %s\n", addr)
		}(addr)
	}
}

// GET request at /view
func getViewHandler(w http.ResponseWriter, r *http.Request) {
	viewMutex.Lock()
	defer viewMutex.Unlock()

	// Collect all replicas into a slice
	currentView := make([]string, 0, len(view))
	for replica := range view {
		currentView = append(currentView, replica)
	}

	// Respond with the current view
	response := map[string]interface{}{"view": currentView}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// PUT request at /view
func putViewHandler(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	newReplica, exists := body["socket-address"]
	if !exists || newReplica == "" {
		http.Error(w, "Missing or empty socket-address", http.StatusBadRequest)
		return
	}

	viewMutex.Lock()
	if _, alreadyPresent := view[newReplica]; alreadyPresent {
		viewMutex.Unlock()
		// Replica already exists
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"result": "already present"})
		return
	}

	// Add the new replica to the view
	view[newReplica] = true
	viewMutex.Unlock()

	// Broadcast the updated view
	go broadcastViewUpdate()

	// Respond with success
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"result": "added"})
}

// DELETE request at /view
func deleteViewHandler(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		//Error Handling
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	replica := body["socket-address"]
	if replica == "" {
		//Error handling
		http.Error(w, "Missing socket-address in request body", http.StatusBadRequest)
		return
	}

	viewMutex.Lock()
	if _, exists := view[replica]; !exists {
		viewMutex.Unlock()
		//Replica not found error handling
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "View has no such replica"})
		return
	}

	// Removing replica from the view
	delete(view, replica)
	viewMutex.Unlock()

	//Broadcast updated view to other replicas
	go broadcastViewUpdate()

	// Respond with success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"result": "deleted"})
}

// Detecing if a replica is down
func ReplicaDown() {
	for {
		viewMutex.Lock()
		for replica := range view {
			if replica == ipAddr {
				continue
			}
			// Checking replica status health with a GET request
			resp, err := http.Get("http://" + replica + "/health")
			if err != nil || resp.StatusCode != http.StatusOK {
				//Commenting and removing the replica
				fmt.Printf("Replica %s is down. Removing from view.\n", replica)
				delete(view, replica)
				go broadcastViewUpdate()
			}
		}
		viewMutex.Unlock()
		//Check every 5 seconds
		time.Sleep(5 * time.Second)
	}
}

// Checkpoint that a server is good to go
func StatusHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// END OF VIEW

// PUT request at /kvs/<key> with JSON body
func PutRequest(w http.ResponseWriter, r *http.Request) {

	//Response Body
	response := make(map[string]interface{})

	//Retrieving the <key>
	key := r.URL.Path[len("/kvs/"):]

	//400 (Bad Request) {"error": "Key is too long"}
	if len(key) > 50 {
		response["error"] = "Key is too long"
		send_r, _ := json.Marshal(response)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(send_r)
		return
	}

	//Retrieving the <value>
	var value map[string]interface{}
	rawJson, _ := io.ReadAll(r.Body)
	json.Unmarshal(rawJson, &value)

	//Getting Causal MetaData Out (causal consistency)
	metadata, has_metadata := value["causal-metadata"].(map[string]interface{})
	var versionReq int
	var has_version bool
	if has_metadata {
		ver_2, hs := metadata[key].(float64)
		has_version = hs
		versionReq = int(ver_2)
	}

	//503 (Service Unavailable) {"error": "Causal dependencies not satisfied; try again later"}
	var versionServer = stringToIntMap[key]
	if has_version {
		if versionReq != versionServer { //checked in test_assignment3 (line 374)
			response["error"] = "Causal dependencies not satisfied; try again later"
			send_r, _ := json.Marshal(response)
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write(send_r)
			return
		}
	} else {
		fmt.Println("No version found")
	}

	valueAttribute, doesItExists := value["value"]

	//400 (Bad Request) {"error": "PUT request does not specify a value"}
	if !doesItExists {
		response["error"] = "PUT request does not specify a value"
		send_r, _ := json.Marshal(response)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(send_r)
		return
	}

	//incrementing the version var
	stringToIntMap[key]++

	// 201 (Created) {"result": "created", "causal-metadata": <V'>}
	if _, exists := kvs[key]; !exists {
		kvs[key] = valueAttribute
		response["result"] = "created"
		response["causal-metadata"] = stringToIntMap
		send_r, _ := json.Marshal(response)
		w.WriteHeader(http.StatusCreated)
		w.Write(send_r)
	} else { //200 (Ok) {"result": "replaced", "causal-metadata": <V'>}
		kvs[key] = valueAttribute
		response["result"] = "replaced"
		response["causal-metadata"] = stringToIntMap
		send_r, _ := json.Marshal(response)
		w.WriteHeader(http.StatusOK)
		w.Write(send_r)
	}

	go PutReplication(key, valueAttribute.(string), stringToIntMap[key])
}

// GET request at /kvs/<key> with JSON body
func GetRequest(w http.ResponseWriter, r *http.Request) {
	response := make(map[string]interface{})
	key := r.URL.Path[len("/kvs/"):]

	// Retrieve the version from the request
	var value map[string]interface{}
	rawJson, _ := io.ReadAll(r.Body)
	json.Unmarshal(rawJson, &value)

	metadata, hasMetadata := value["causal-metadata"].(map[string]interface{})
	var versionReq int
	if hasMetadata {
		if ver, ok := metadata[key].(float64); ok {
			versionReq = int(ver)
		}
	}

	// Check if the key exists in the local KVS
	if entryValue, exists := kvs[key]; exists {
		versionServer := stringToIntMap[key]

		// Check causal metadata
		if hasMetadata && versionReq > versionServer {
			response["error"] = "Causal dependencies not satisfied; try again later"
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Key found
		response["result"] = "found"
		response["value"] = entryValue
		response["causal-metadata"] = stringToIntMap
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	} else {
		// Key not found
		response["error"] = "Key does not exist"
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
	}
}

// DELETE request at /kvs/<key> with JSON body
func DeleteRequest(w http.ResponseWriter, r *http.Request) {

	response := make(map[string]interface{})
	key := r.URL.Path[len("/kvs/"):] // Extracting the key

	//Retrieving the version
	var requestBody map[string]interface{}
	rawJson, _ := io.ReadAll(r.Body)
	json.Unmarshal(rawJson, &requestBody)

	//Getting Causal MetaData Out (causal consistency)
	metadata, has_metadata := requestBody["causal-metadata"].(map[string]interface{})
	var versionReq int
	var has_version bool
	if has_metadata {
		ver_2, hs := metadata[key].(float64)
		has_version = hs
		versionReq = int(ver_2)
	}

	//503 (Service Unavailable) {"error": "Causal dependencies not satisfied; try again later"}
	var versionServer = stringToIntMap[key]
	if has_version {
		if versionReq != versionServer { //checked in test_assignment3 (line 374)
			response["error"] = "Causal dependencies not satisfied; try again later"
			send_r, _ := json.Marshal(response)
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write(send_r)
			return
		}
	} else {
		fmt.Println("No version found")
	}

	// 404 (Not Found) {"error": "Key does not exist"}
	if _, exists := kvs[key]; !exists {
		response["error"] = "Key does not exist"
		send_r, _ := json.Marshal(response)
		w.WriteHeader(http.StatusNotFound)
		w.Write(send_r)
		return
	}

	// Key exists, proceed with deletion
	delete(kvs, key)
	response["result"] = "deleted"

	//incrementing the version var
	stringToIntMap[key]++

	response["causal-metadata"] = stringToIntMap
	send_r, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(send_r)

	// Replicate deletion to other replicas
	go DeleteReplication(key, stringToIntMap[key])
}

// Method Handler
func methodHandler(w http.ResponseWriter, r *http.Request) {
	method := r.Method

	switch method {
	case "GET":
		GetRequest(w, r)
		break

	case "PUT":
		PutRequest(w, r)
		break

	case "DELETE":
		DeleteRequest(w, r)
		break

	default:
		return
	}
}

// Replication Handler (when recieving a request)
func replicaHandler(w http.ResponseWriter, r *http.Request) {
	var rec_body map[string]interface{}
	rawJson, _ := io.ReadAll(r.Body)
	json.Unmarshal(rawJson, &rec_body)

	//Retrieving the method, key, rec_body, and version
	method, _ := rec_body["method"]
	key, _ := rec_body["key"].(string)
	value, _ := rec_body["value"]
	versionString, _ := rec_body["version"]
	version := int(versionString.(float64))

	//Check if the recieve version is less than the server version
	stringToIntMap[key] = version

	if method == "PUT" {
		kvs[key] = value
	}
	if method == "DELETE" {
		delete(kvs, key)
	}

	var response map[string]interface{} = map[string]interface{}{}
	send_r, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(send_r)
}

// Main http server
func main() {

	//Initialize the view based on the "VIEW" environment variable
	initializeView()

	http.HandleFunc("/view", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getViewHandler(w, r)
		case http.MethodPut:
			putViewHandler(w, r)
		case http.MethodDelete:
			deleteViewHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	//Register routes for key-value operations, replication, cluster view management, view, status health
	http.HandleFunc("/view/update", updateViewHandler)
	http.HandleFunc("/kvs/", methodHandler)
	http.HandleFunc("/kvs/sync", syncHandler)
	http.HandleFunc("/replica", replicaHandler)
	http.HandleFunc("/health", StatusHealth)

	//Detecting replicas that are down
	go ReplicaDown()

	fmt.Fprintln(os.Stdout, "Server running!")
	http.ListenAndServe(ipAddr, nil)
}
