# terraform-ibm-provider-api

It is used to perform terraform action on ibm cloud provider using REST API

# Dependencies

-	[Terraform](https://www.terraform.io/downloads.html) 0.9.3+
-	[Go](https://golang.org/doc/install) 1.8 (to build the provider plugin)
-   [IBM Cloud Provider](https://github.com/IBM-Cloud/terraform-provider-ibm/)


## Files

*   main.go

    This file contains the web server and handlers.

*   cmd/discovery

    Code for the executable.

## Steps to use the project as an executable

*  Build and install the executable to your GOPATH

       make install-cli

*  Example commands

       discovery config --git_url https://github.com/srikar-git/ibm-vsi --config_dir testi
       discovery import --services ibm_is_vpc --config_dir testi --repo_name ibm-vsi


## Steps to use the project as a server

*  Start the server

       export MOUNT_DIR=<dir to clone the repo>
       go run main.go docs.go

*  Or 

       make run-mac <or make run-local for linux>

*  Or run as docker container

       make docker-build
       make docker-run

    First two need mongodb service running on localhost:27017. Third needs mongodb running as docker container. To run mongodb as docker container with 27017 exposed outside. This will work for all three methods above. Run this before any of the above steps
        
        make docker-run-mongo
        


## How to run the terraform-ibmcloud-provider-api as a container
        
        cd /go/src/github.com
        git clone git@github.ibm.com:IBMTerraform/terraform-provider-ibm-api/.git
        cd terraform-provider-ibm-api/
        docker build -t terraform-provider-ibm-api .
        docker images
        export API_IMAGE=terraform-provider-ibm-api:latest
        export MOUNT_DIR=<dir where repo will be cloned>
        docker-compose up --build -d
        
*  Create the configuration <br />
     
        URL: http://<HOST>:9080/configuration
        METHOD: POST 
        HEADER: 
          Content-Type: application/json
          Accept: application/json
        SAMPLE Payload for configuration:
            {  
                // Provide git url for your terraform configuration git repo.
                "git_url":"https://github.com/sakshiag/speech-to-text-terraform",

                // Provide the variable required to run the configuration.
                "variablestore":[  
                {  
                    "name":"org",
                    "value":"org_name"
                },
                {  
                    "name":"space",
                    "value":"space_name"
                },
                {       
                    "name":"region",
                    "value":"region"
                },
                {  
                    "name":"datacenter",
                    "value":"datcenter"
                },
                {  
                    "name":"machine_type",
                    "value":"machine_type"
                },
                {  
                    "name":"isolation",
                    "value":"public"
                },
                {  
                    "name":"private_vlan_id",
                    "value":"private_vlan_id"
                },
                {  
                    "name":"public_vlan_id",
                    "value":"<public_vlan_id>"
                },
                {  
                    "name":"subnet_id",
                    "value":"subent_id"
                },
                {  
                    "name":"bluemix_api_key",
                    "value":"bm_api_key"
                }],
                // To define the terraform log level It is optional
                "log_level": "DEBUG"
            }

        Response:
            {
                "id": <conig name is returned>
            }

* Perform the action (apply, plan and delete) <br />

        //config_id is the id returned from /configuration API.
        //action can be PLAN,APPLY,DELETE and SHOW.
        URL: http://<HOST>:9080/configuration/config_id/{action}
        METHOD: POST
        HEADER: 
          Content-Type: application/json
          Accept: application/json
          SLACK_WEBHOOK_URL: <provide your slack webhook url.>
        Response:
            {
                "id": <action_id is returned which is used to retrive the logs and status.>,
            }

* Get the status of the action <br />

        //config_id is the id returned from /configuration API.
        //action_id can be PLAN,APPLY,DELETE and SHOW.
        URL: http://<HOST>:9080/configuration/config_id/{action}/{action_id}/status
        METHOD: GET
        HEADER: 
          Content-Type: application/json
          Accept: application/json
        Response:
            {
                "status" : <status of the action>,
                "error" : <error if any error occured.>
            }

* Get the logs of the action <br />

        //config_id is the id returned from /configuration API.
        //action_id can be PLAN,APPLY,DELETE and SHOW.
        URL: http://<HOST>:9080/configuration/config_id/{action}/{action_id}/log
        METHOD: GET
        HEADER: 
          Content-Type: application/json
          Accept: application/json
        Response:
            {
                "action" : "action_name",
                "id" : "action_id",
                "output" : "output logs",
                "error" : "error logs"
            }

* Delete the configuration. <br />

        //config_id is the id returned from /configuration API.
        URL: http://<HOST>:9080/configuration/config_id
        METHOD: DELETE
        HEADER: 
          Content-Type: application/json
          Accept: application/json
        Response: 200 OK
