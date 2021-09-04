# IBM Cloud Configuration Discovery

Use "Configuration Discovery" to import the existing Cloud resources (in your account) and its configuration settings - to auto-generate the terraform configuration file (.tf) and state file (.tfstate).  It makes it easy for you to adopt the Infrastructure-as-Code practices; it can reverse engineer the current IBM Cloud environment (that was provisioned using UI or CLI).  

# Dependencies

-	[Terraform](https://www.terraform.io/downloads.html) 0.9.3+
-	[Terraformer](https://github.com/GoogleCloudPlatform/terraformer) 0.8.15+
-	[Go](https://golang.org/doc/install) 1.15+ (to build the provider plugin)
-   [IBM Cloud Provider](https://github.com/IBM-Cloud/terraform-provider-ibm/)
-   [Mongodb](https://docs.mongodb.com/manual/installation/) v4.4.5+


## Steps to use the Configuration Discovery project
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

        cd /go/src/github.com
        git clone git@github.ibm.com:IBMTerraform/configuration-discovery/.git
        cd configuration-discovery/
        go run main.go docs.go
        http://localhost:8080

## How to run the Configuration Discovery as a docker container
*  Or 

       make run-mac <or make run-local for linux>

*  Or run as docker container

       make docker-build
       make docker-run

    First two need mongodb service running on localhost:27017. Third needs mongodb running as docker container. To run mongodb as docker container with 27017 exposed outside. This will work for all three methods above. Run this before any of the above steps
        
        make docker-run-mongo
        


## How to run the terraform-ibmcloud-provider-api as a container
        
        cd /go/src/github.com
        git clone git@github.ibm.com:IBMTerraform/configuration-discovery/.git
        cd configuration-discovery/
        docker build -t configuration-discovery .
        docker images
        export API_IMAGE=configuration-discovery:latest
        docker-compose up --build -d
        
## Contributing to Configuration Discovery

Please have a look at the [CONTRIBUTING.md](./CONTRIBUTING.md) file for some guidance before
submitting a pull request. Thank you for your help and interest!

#### Report a Issue / Feature request

-   Is something broken? Have a issue/bug to report? use the [Bug report](https://github.com/IBM-Cloud/configuration-discovery/issues/new?assignees=&labels=&template=bug_report.md&title=) link. But before raising a issue, please check the [issues list](https://github.com/IBM-Cloud/configuration-discovery/issues) to see if the issue is already raised by someone
-   Do you have a new feature or enhancement you would like to see? use the [Feature request](https://github.com/IBM-Cloud/configuration-discovery/issues/new?assignees=&labels=&template=feature_request.md&title=) link.

## License

The project is licensed under the [Apache 2.0 License](https://www.apache.org/licenses/LICENSE-2.0).
A copy is also available in the LICENSE file in this repository.