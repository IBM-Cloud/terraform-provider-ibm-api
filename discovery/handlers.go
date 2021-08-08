// @SubApi ibmcloud-terraform-provider [/v1]
// @SubApi Allows you access ibm cloud terraform provider api [/v1]

package discovery

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/IBM-Cloud/terraform-provider-ibm-api/utils"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	mgo "gopkg.in/mgo.v2"
)

var httpClient *http.Client
var sessionMgo *mgo.Session
var githubToken string
var githubIBMToken string
var planTimeOut = 60 * time.Minute
var currentOps = make(map[string]chan utils.StatusResponse)

// VariablesRequest -
type VariablesRequest []EnvironmentVariableRequest

// EnvironmentVariableRequest -
type EnvironmentVariableRequest struct {
	Name  string `json:"name,required" binding:"required" description:"The variable's name"`
	Value string `json:"value,required" binding:"required" description:"The variable's value"`
}

//ConfHandler handles request to kickoff git clone of the repo.
// @Title ConfHandler
// @Description clone the configuration repo
// @Accept  json
// @Produce  json
// @Param   body     body     ConfigRequest   true "request body"
// @Success 200 {object} ConfigResponse
// @Failure 500 {object} string
// @Failure 400 {object} string
// @Router /v2/configuration [post]
func ConfHandler(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// Read body
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// Unmarshal
		var msg utils.ConfigRequest
		var response utils.ConfigResponse
		var configName string
		err = json.Unmarshal(b, &msg)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		if msg.GitURL == "" && msg.ConfigName == "" {
			//Create discovery directory to import tf/state file of services
			configName = "discovery"
			err = createDir(utils.GetConfiguration().Server.MountDir + "/" + configName)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
		} else if msg.ConfigName != "" {
			//Create config directory to import tf/state file of discovery services
			configName = msg.ConfigName
			err = createDir(utils.GetConfiguration().Server.MountDir + "/" + configName)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
		} else {
			log.Println(msg.GitURL)
			log.Println("Will clone git repo")
			_, configName, err := utils.CloneRepo(msg)
			if err != nil {
				log.Println("Eror Cloning repo..")
				log.Printf("err : %v\n", err)
				return
			}
			log.Println("\n", configName)
		}

		if msg.LOGLEVEL != "" {
			os.Setenv("TF_LOG", msg.LOGLEVEL)
		}

		response.ConfigName = configName
		log.Println(response)

		output, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return
		}

		w.Header().Set("content-type", "application/json")
		w.Write(output)
	}
}

//TerraformerImportHandler handles request to get the terraform resources & state file.
// @Title TerraformerImportHandler
// @Description Get status of the action.
// @Param   repo_name   path     string      true "repo name"
// @Param   service     query    string     true "service"
// @Accept  json
// @Produce  json
// @Success 200 {object} StatusResponse
// @Failure 404 {object} string
// @Failure 500 {object} string
// @Router /v2/configuration/{repo_name}/import [POST]
func TerraformerImportHandler(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		var actionResponse utils.ActionResponse
		var statusResponse utils.StatusResponse

		// Read body
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		// Read Query Parameter
		configName := r.URL.Query().Get("repo_name")
		services := r.URL.Query().Get("services")
		command := r.URL.Query().Get("command")
		tags := r.URL.Query().Get("tags")
		opts := []string{}
		func() {
			if services != "" {
				opts = append(opts, "--resources="+services)
			}
			if tags != "" {
				splittedTags := strings.Split(tags, ",")
				fmt.Println(splittedTags)
				if len(splittedTags) > 0 {
					for _, v := range splittedTags {
						tag := strings.SplitN(v, ":", 2)
						if len(tag) == 2 {
							opts = append(opts, fmt.Sprintf("--%s=%s", strings.TrimSpace(strings.ToLower(tag[0])), tag[1]))
						}
					}
				}
			}
		}()

		b = make([]byte, 10)
		rand.Read(b)
		randomID := fmt.Sprintf("%x", b)

		//Clean up discovery directory
		discoveryDir := utils.GetConfiguration().Server.MountDir + "/" + "discovery"
		err = removeDir(discoveryDir + "/*")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		go func() {
			if command == "default" {
				if configName != "discovery" {
					discoveryDir = utils.GetConfiguration().Server.MountDir + "/" + configName
				}
				err = DiscoveryImport(configName, randomID, discoveryDir, opts)
				if err != nil {
					statusResponse.Error = err.Error()
					statusResponse.Status = "Failed"

					// Update the status in the db in case it is failed
					err = utils.UpdateMongodb(s, randomID, statusResponse.Status)
					if err != nil {
						http.Error(w, err.Error(), 500)
						return
					}
					return
				}
				statusResponse.Status = "Completed"
				// Update the status in the db in case it is completed
				err = utils.UpdateMongodb(s, randomID, statusResponse.Status)
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
			} else if command == "merge" {
				err = DiscoveryImport(configName, randomID, discoveryDir, opts)
				if err != nil {
					statusResponse.Error = err.Error()
					statusResponse.Status = "Failed"

					// Update the status in the db in case it is failed
					err = utils.UpdateMongodb(s, randomID, statusResponse.Status)
					if err != nil {
						http.Error(w, err.Error(), 500)
						return
					}
					return
				}

				//Merge state files and templates
				repoDir := utils.GetConfiguration().Server.MountDir + "/" + configName
				//Backup repo TF file.
				err = copy(repoDir+"/terraform.tfstate", repoDir+"/terraform.tfstate_backup")
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}

				//Read state file from local repo directory
				terraformStateFile := repoDir + "/terraform.tfstate"
				terraformObj := ReadTerraformStateFile(terraformStateFile, "")

				//Read state file from discovery repo directory
				terraformerSateFile := discoveryDir + "/terraform.tfstate"
				terraformerObj := ReadTerraformStateFile(terraformerSateFile, "discovery")

				// comparing state files
				if cmp.Equal(terraformObj, terraformerObj, cmpopts.IgnoreFields(Resource{}, "ResourceName")) {
					log.Printf("# Config repo configuration/state is equal !!\n")
				} else {
					log.Printf("# Config repo configuration/state is not equal !!\n")
					err = MergeStateFile(terraformObj, terraformerObj, terraformerSateFile, terraformStateFile, repoDir, "", randomID, &planTimeOut)
					if err != nil {
						http.Error(w, err.Error(), 500)
						return
					}
				}
				statusResponse.Status = "Completed"
				// Update the status in the db in case it is completed
				err = utils.UpdateMongodb(s, randomID, statusResponse.Status)
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
			}
		}()

		if command != "merge" && command != "default" {
			errMsg := "command value not supported. Please provide 'default' or 'merge' as command value!!"
			log.Printf("# '%s' %s ", command, errMsg)

			w.WriteHeader(500)
			actionResponse.Status = errMsg
		} else {
			w.WriteHeader(200)
			actionResponse.Status = "In-Progress"
		}

		actionResponse.Action = "import"
		actionResponse.ConfigName = configName
		actionResponse.ActionID = randomID
		actionResponse.Timestamp = time.Now().Format("20060102150405")

		// Make an entry in the db
		utils.InsertMongodb(s, actionResponse)

		output, err := json.MarshalIndent(actionResponse, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("content-type", "application/json")
		w.Write(output)

	}
}

//TerraformerStateHandler handles request to get the terraform resources & state file.
// @Title TerraformerImportHandler
// @Description Get status of the action.
// @Param   repo_name   path     string      true "repo name"
// @Param   service     query    string     true "service"
// @Accept  json
// @Produce  json
// @Success 200 {object} StatusResponse
// @Failure 404 {object} string
// @Failure 500 {object} string
// @Router /v2/configuration/{repo_name}/import [GET]
func TerraformerStateHandler(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		var actionResponse utils.ActionResponse

		// Read body
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		// Read Query Parameter
		services := r.URL.Query().Get("service")
		configName := "terraformer"
		confDir := path.Join(utils.GetConfiguration().Server.MountDir, configName)

		b = make([]byte, 10)
		rand.Read(b)
		randomID := fmt.Sprintf("%x", b)

		go func() {

			//Merge state files and templates
			b = make([]byte, 10)
			rand.Read(b)

			s := strings.Split(services, ",")
			if len(s) > 0 {
				for _, srv := range s {
					srvDir := confDir + "/generated" + "/ibm/" + srv

					//Backup TF file.
					err = copy(srvDir+"/terraform.tfstate_backup", srvDir+"/terraform.tfstate")
					if err != nil {
						http.Error(w, err.Error(), 500)
						return
					}

				}
			}

		}()

		w.WriteHeader(200)

		actionResponse.Action = "statefile"
		actionResponse.ConfigName = configName
		actionResponse.ActionID = randomID
		actionResponse.Timestamp = time.Now().Format("20060102150405")
		actionResponse.Status = "Completed"

		// Make an entry in the db
		utils.InsertMongodb(s, actionResponse)

		output, err := json.MarshalIndent(actionResponse, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("content-type", "application/json")
		w.Write(output)

	}
}
