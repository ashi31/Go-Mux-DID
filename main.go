package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	srvcfg "github.com/EnsurityTechnologies/config"
	"github.com/EnsurityTechnologies/logger"
	"github.com/gorilla/mux"
	"github.com/rubixchain/rubixgoplatform/client"
	"github.com/rubixchain/rubixgoplatform/did"
)

type DIDCreate struct {
	Type              int    `json:"type"`
	Dir               string `json:"dir"`
	Config            string `json:"config"`
	MasterDID         string `json:"master_did"`
	Secret            string `json:"secret"`
	PrivPWD           string `json:"priv_pwd"`
	QuorumPWD         string `json:"quorum_pwd"`
	ImgFile           string `json:"img_file"`
	DIDImgFileName    string `json:"did_img_file"`
	PubImgFile        string `json:"pub_img_file"`
	PrivImgFile       string `json:"priv_img_file"`
	PubKeyFile        string `json:"pub_key_file"`
	PrivKeyFile       string `json:"priv_key_file"`
	QuorumPubKeyFile  string `json:"quorum_pub_key_file"`
	QuorumPrivKeyFile string `json:"quorum_priv_key_file"`
}

type DIDCreationResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Result  struct {
		DID    string `json:"did"`
		PeerID string `json:"peer_id"`
	} `json:"result"`
}

func createDIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var didConfig did.DIDCreate
	err := r.ParseMultipartForm(10 << 20) // 10 MB maximum request size
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusInternalServerError)
		return
	}

	didConfigData := r.FormValue("did_config")
	err = json.Unmarshal([]byte(didConfigData), &didConfig)
	if err != nil {
		http.Error(w, "Error parsing JSON data", http.StatusBadRequest)
		return
	}

	response, err := forwardRequestToOtherServer(&didConfig)
	if err != nil {
		http.Error(w, "Error forwarding request", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)

}

func forwardRequestToOtherServer(config *did.DIDCreate) (*DIDCreationResponse, error) {

	fp, err := os.OpenFile("logFile.txt",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	logOptions := &logger.LoggerOptions{
		Name:   "Main",
		Level:  3,
		Color:  []logger.ColorOption{logger.AutoColor, logger.ColorOff},
		Output: []io.Writer{logger.DefaultOutput, fp},
	}
	log := logger.New(logOptions)
	timeout := time.Duration(0) * time.Minute

	client, err := client.NewClient(&srvcfg.Config{ServerAddress: "localhost", ServerPort: "20001"}, log, timeout)

	if err != nil {
		println(err)
	}

	did, done := client.CreateDID(config)

	println(did)
	println(done)

	var responseBody DIDCreationResponse

	responseBody.Status = done
	responseBody.Message = "DID is Created"
	responseBody.Result.DID = did

	return &responseBody, nil
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/api/createdid", createDIDHandler).Methods("POST")

	fmt.Println("Server started on :20000")
	http.ListenAndServe(":20000", router)
}
