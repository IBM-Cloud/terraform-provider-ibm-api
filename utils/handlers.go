package utils

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/gorilla/mux"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var httpClient *http.Client
var sessionMgo *mgo.Session
var githubToken string
var githubIBMToken string
var planTimeOut = 60 * time.Minute
var currentOps = make(map[string]chan StatusResponse)

// Message -
type Message struct {
	GitURL        string            `json:"git_url,required" description:"The git url of your configuraltion"`
	VariableStore *VariablesRequest `json:"variablestore,omitempty" description:"The environments' variable store"`
	LOGLEVEL      string            `json:"log_level,omitempty" description:"The log level defing by user."`
}

// ConfigResponse -
type ConfigResponse struct {
	ID string `json:"id,required" description:"ID of the git operation."`
}

// StatusResponse -
type StatusResponse struct {
	Status string `json:"status,required" description:"Status of the terraform operation."`
	Error  string `json:"error,omitemplty" description:"Error of the terraform operation."`
}

// ActionResponse -
type ActionResponse struct {
	ConfigName string `json:"id,required" description:"Name of the configuration"`
	Action     string `json:"action,required" description:"Action Name"`
	ActionID   string `json:"action_id"`
	Timestamp  string `json:"timestamp"`
	Status     string `json:"status"`
}

// ActionDetails -
type ActionDetails struct {
	ConfigName string `json:"id,required" description:"Name of the configuration"`
	Action     string `json:"action,required" description:"Action Name"`
	ActionID   string `json:"action_id"`
	Output     string `json:"output"`
	Error      string `json:"error"`
}

// VariablesRequest -
type VariablesRequest []EnvironmentVariableRequest

// EnvironmentVariableRequest -
type EnvironmentVariableRequest struct {
	Name  string `json:"name,required" binding:"required" description:"The variable's name"`
	Value string `json:"value,required" binding:"required" description:"The variable's value"`
}

var currentDir = os.Getenv("MOUNT_DIR")

var logDir = currentDir + "/log/"

var stateDir = currentDir + "/state"

func init() {

	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		os.MkdirAll(logDir, os.ModePerm)
	}
	if _, err := os.Stat(stateDir); os.IsNotExist(err) {
		os.MkdirAll(stateDir, os.ModePerm)
	}

}

//ConfHandler handles request to kickoff git clone of the repo.
func ConfHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method.", 405)
	}

	// Read body
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Unmarshal
	var msg Message
	var response ConfigResponse
	err = json.Unmarshal(b, &msg)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	log.Println(msg.GitURL)
	if msg.GitURL == "" {
		w.WriteHeader(400)
		w.Write([]byte("EMPTY GIT URL"))
		return
	}

	if msg.LOGLEVEL != "" {
		os.Setenv("TF_LOG", msg.LOGLEVEL)
	}

	log.Println("Will clone git repo")

	_, id, err := cloneRepo(msg)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	log.Println("\n", id)

	response.ID = id
	log.Println(response)

	output, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return
	}

	confDir := path.Join(currentDir, id)

	b = make([]byte, 10)
	rand.Read(b)
	randomID := fmt.Sprintf("%x", b)

	err = TerraformInit(confDir, id, &planTimeOut, randomID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)
}

//ConfDeleteHandler handles request to kickoff git clone of the repo.
func ConfDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Invalid request method.", 405)
	}

	vars := mux.Vars(r)
	repoName := vars["repo_name"]

	err := removeRepo(currentDir, repoName)
	if err != nil {
		w.WriteHeader(404)
		log.Println(err)
		w.Write([]byte(fmt.Sprintf("There is no config repo file for this request.")))
		return
	}
}

//PlanHandler handles request to run terraform plan.
func PlanHandler(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		webhook := r.Header.Get("SLACK_WEBHOOK_URL")
		vars := mux.Vars(r)
		repoName := vars["repo_name"]

		var actionResponse ActionResponse
		var statusResponse StatusResponse

		log.Println("Url Param 'repo name' is: " + repoName)
		confDir := path.Join(currentDir, repoName)

		b := make([]byte, 10)
		rand.Read(b)
		randomID := fmt.Sprintf("%x", b)
		commandDone := make(chan StatusResponse)
		currentOps[randomID] = commandDone

		outURL := "http://" + r.Host + "/" + r.URL.Path + "/" + randomID + ".out"
		errURL := "http://" + r.Host + "/" + r.URL.Path + "/" + randomID + ".err"

		// Post to slack that the action has started and the link logs
		ResultToSlack(outURL, errURL, "plan", randomID, "In-Progress", webhook)

		go func() {
			pullRepo(repoName)
			err := TerraformPlan(confDir, repoName, &planTimeOut, randomID)
			if err != nil {
				statusResponse.Error = err.Error()
				statusResponse.Status = "Failed"
				commandDone <- statusResponse

				// Update the status in the db in case it is failed
				err = UpdateMongodb(s, randomID, statusResponse.Status)
				return
			}
			statusResponse.Status = "Completed"
			commandDone <- statusResponse

			// Update the status in the db in case it is completed
			err = UpdateMongodb(s, randomID, statusResponse.Status)
			ResultToSlack(outURL, errURL, "plan", randomID, "Completed", webhook)
		}()

		w.WriteHeader(202)

		actionResponse.Action = "plan"
		actionResponse.ConfigName = repoName
		actionResponse.ActionID = randomID
		actionResponse.Timestamp = time.Now().Format("20060102150405")
		actionResponse.Status = "In-Progress"

		// Make an entry in the db
		insertMongodb(s, actionResponse)

		output, err := json.MarshalIndent(actionResponse, "", "  ")
		if err != nil {
			return
		}
		w.Header().Set("content-type", "application/json")
		w.Write(output)

	}
}

//ApplyHandler handles request to run terraform plan.
func ApplyHandler(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		webhook := r.Header.Get("SLACK_WEBHOOK_URL")

		var actionResponse ActionResponse
		var statusResponse StatusResponse

		vars := mux.Vars(r)
		repoName := vars["repo_name"]

		log.Println("Url Param 'repo name' is: " + repoName)
		confDir := path.Join(currentDir, repoName)

		b := make([]byte, 10)
		rand.Read(b)
		randomID := fmt.Sprintf("%x", b)
		commandDone := make(chan StatusResponse)
		currentOps[randomID] = commandDone
		outURL := "http://" + r.Host + "/" + r.URL.Path + "/" + randomID + ".out"
		errURL := "http://" + r.Host + "/" + r.URL.Path + "/" + randomID + ".err"
		ResultToSlack(outURL, errURL, "apply", randomID, "In-Progress", webhook)
		go func() {

			pullRepo(repoName)
			err := TerraformApply(confDir, stateDir, repoName, &planTimeOut, randomID)
			if err != nil {
				statusResponse.Error = err.Error()
				statusResponse.Status = "Failed"
				commandDone <- statusResponse
				err = UpdateMongodb(s, randomID, statusResponse.Status)
				return
			}
			statusResponse.Status = "Completed"
			ResultToSlack(outURL, errURL, "apply", randomID, statusResponse.Status, webhook)
			err = UpdateMongodb(s, randomID, statusResponse.Status)
			commandDone <- statusResponse
		}()
		w.WriteHeader(202)
		actionResponse.Action = "apply"
		actionResponse.ConfigName = repoName
		actionResponse.ActionID = randomID
		actionResponse.Timestamp = time.Now().Format("20060102150405")
		actionResponse.Status = "In-Progress"

		insertMongodb(s, actionResponse)
		output, err := json.MarshalIndent(actionResponse, "", "  ")
		if err != nil {
			return
		}
		w.Header().Set("content-type", "application/json")
		w.Write(output)

	}
}

//DestroyHandler handles request to run terraform plan.
func DestroyHandler(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		var actionResponse ActionResponse
		var statusResponse StatusResponse

		webhook := r.Header.Get("SLACK_WEBHOOK_URL")
		vars := mux.Vars(r)
		repoName := vars["repo_name"]

		log.Println("Url Param 'repo name' is: " + repoName)
		confDir := path.Join(currentDir, repoName)

		b := make([]byte, 10)
		rand.Read(b)
		randomID := fmt.Sprintf("%x", b)
		commandDone := make(chan StatusResponse)
		currentOps[randomID] = commandDone
		outURL := "http://" + r.Host + "/" + r.URL.Path + "/" + randomID + ".out"
		errURL := "http://" + r.Host + "/" + r.URL.Path + "/" + randomID + ".err"
		ResultToSlack(outURL, errURL, "destroy", randomID, "In-Progress", webhook)
		go func() {
			err := TerraformDestroy(confDir, stateDir, repoName, &planTimeOut, randomID)
			if err != nil {
				statusResponse.Error = err.Error()
				statusResponse.Status = "Failed"
				commandDone <- statusResponse
				err = UpdateMongodb(s, randomID, statusResponse.Status)
				return
			}
			statusResponse.Status = "Completed"
			ResultToSlack(outURL, errURL, "destroy", randomID, "Completed", webhook)
			err = UpdateMongodb(s, randomID, statusResponse.Status)
			commandDone <- statusResponse

		}()

		w.WriteHeader(202)
		actionResponse.Action = "destroy"
		actionResponse.ConfigName = repoName
		actionResponse.ActionID = randomID
		actionResponse.Timestamp = time.Now().Format("20060102150405")
		actionResponse.Status = "In-Progress"

		insertMongodb(s, actionResponse)

		output, err := json.MarshalIndent(actionResponse, "", "  ")
		if err != nil {
			return
		}
		w.Header().Set("content-type", "application/json")
		w.Write(output)

	}
}

//ShowHandler handles request to run terraform show.
func ShowHandler(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		session := s.Copy()
		defer session.Close()

		var actionResponse ActionResponse
		var statusResponse StatusResponse

		webhook := r.Header.Get("SLACK_WEBHOOK_URL")
		vars := mux.Vars(r)

		repoName := vars["repo_name"]

		log.Println("Url Param 'repo name' is: " + repoName)
		confDir := path.Join(currentDir, repoName)

		b := make([]byte, 10)
		rand.Read(b)
		randomID := fmt.Sprintf("%x", b)
		commandDone := make(chan StatusResponse)
		currentOps[randomID] = commandDone
		outURL := "http://" + r.Host + "/" + r.URL.Path + "/" + randomID + ".out"
		errURL := "http://" + r.Host + "/" + r.URL.Path + "/" + randomID + ".err"
		ResultToSlack(outURL, errURL, "show", randomID, "In-Progress", webhook)
		go func() {
			err := TerraformShow(confDir, stateDir, repoName, &planTimeOut, randomID)
			if err != nil {
				statusResponse.Error = err.Error()
				statusResponse.Status = "Failed"
				commandDone <- statusResponse
				err = UpdateMongodb(s, randomID, statusResponse.Status)
				return
			}
			outURL := "http://" + r.Host + "/" + r.URL.Path + "/" + randomID + ".out"
			errURL := "http://" + r.Host + "/" + r.URL.Path + "/" + randomID + ".err"

			statusResponse.Status = "Completed"
			ResultToSlack(outURL, errURL, "show", randomID, statusResponse.Status, webhook)
			err = UpdateMongodb(s, randomID, statusResponse.Status)
			commandDone <- statusResponse

		}()
		w.WriteHeader(202)

		actionResponse.Action = "show"
		actionResponse.ConfigName = repoName
		actionResponse.ActionID = randomID
		actionResponse.Timestamp = time.Now().Format("20060102150405")
		actionResponse.Status = "In-Progress"

		insertMongodb(s, actionResponse)

		output, err := json.MarshalIndent(actionResponse, "", "  ")
		if err != nil {
			return
		}
		w.Header().Set("content-type", "application/json")
		w.Write(output)

	}
}

//LogHandler handles request to run terraform plan.
func LogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Invalid request method.", 405)
		return
	}
	var response ActionDetails

	vars := mux.Vars(r)
	repoName := vars["repo_name"]
	action := vars["action"]
	actionID := vars["actionID"]

	log.Println("Url Param 'repo name' is: " + repoName)
	log.Println("Url Param 'action' is: " + action)
	log.Println("Url Param 'actionID' is: " + actionID)

	outFile, errFile, err := readLogFile(actionID)

	response.ConfigName = repoName
	response.Output = outFile
	response.Error = errFile
	response.Action = action
	response.ActionID = actionID

	output, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)

}

//StatusHandler handles request to run terraform plan.
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Invalid request method.", 405)
		return
	}
	var response StatusResponse

	vars := mux.Vars(r)
	repoName := vars["repo_name"]
	action := vars["action"]
	actionID := vars["actionID"]

	log.Println("Url Param 'repo name' is: " + repoName)
	log.Println("Url Param 'action' is: " + action)
	log.Println("Url Param 'actionID' is: " + actionID)

	select {
	case value := <-currentOps[actionID]:

		response = value

		go func() {
			currentOps[actionID] <- value
		}()
	default:
		response.Status = "In-Progress"
	}
	output, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)

}

//ViewLogHandler handles request to retrieve the log file
func ViewLogHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, "Invalid request method.", 405)
		return
	}
	vars := mux.Vars(r)
	logFile := vars["log_file"]

	body, err := ioutil.ReadFile(path.Join(logDir, logFile))
	if err != nil {
		w.WriteHeader(404)
		log.Println(err)
		w.Write([]byte(fmt.Sprintf("There is no log file for this request")))
		return
	}
	w.WriteHeader(200)
	w.Write(body)
}

//StateHandler handles request to retrieve the state file
/*func StateHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, "Invalid request method.", 405)
		return
	}
	vars := mux.Vars(r)
	repoName := vars["repo_name"]

	body, err := ioutil.ReadFile(path.Join(stateDir, repoName+".tfstate"))
	if err != nil {
		w.WriteHeader(404)
		log.Println(err)
		w.Write([]byte(fmt.Sprintf("There is no state file for this request")))
		return
	}
	w.WriteHeader(200)
	w.Write(body)
}*/

//GetActionDetailsHandler handles request to run terraform plan.
func GetActionDetailsHandler(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		vars := mux.Vars(r)
		//repoName := vars["repo_name"]
		action := vars["action"]

		var actionResponse []ActionResponse
		c := session.DB("action").C("actionDetails")

		err := c.Find(bson.M{"action": action}).All(&actionResponse)

		w.WriteHeader(202)

		output, err := json.MarshalIndent(actionResponse, "", "  ")

		if err != nil {
			return
		}
		w.Header().Set("content-type", "application/json")
		w.Write(output)

	}
}

//UpdateMongodb updates the status of the action.
func UpdateMongodb(s *mgo.Session, actionID string, status string) error {
	session := s.Copy()
	defer session.Close()
	c := session.DB("action").C("actionDetails")
	err := c.Update(bson.M{"actionid": actionID}, bson.M{"$set": bson.M{"status": status}})
	if err != nil {
		return err
	}

	return nil
}

//UpdateMongodb updates the status of the action.
func insertMongodb(s *mgo.Session, actionResponse ActionResponse) {
	session := s.Copy()
	defer session.Close()
	c := session.DB("action").C("actionDetails")
	err := c.Insert(actionResponse)
	if err != nil {
		if mgo.IsDup(err) {
			return
		}
		log.Println("Failed insert action details : ", err)
		return
	}
}
