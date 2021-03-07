package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/sjson"
)

// ReadTerraformerStateFile ..
// TF 0.12 compatible
func ReadTerraformerStateFile(terraformerStateFile string) ResourceList {
	var rList ResourceList
	tfData := TerraformSate{}

	tfFile, err := ioutil.ReadFile(terraformerStateFile)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal([]byte(tfFile), &tfData)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(tfData.Modules); i++ {
		rData := Resource{}
		for k := range tfData.Modules[i].Resources {
			rData.ResourceName = k
			rData.ResourceType = tfData.Modules[i].Resources[k].ResourceType
			for p := range tfData.Modules[i].Resources[k].Primary {
				if p == "attributes" {
					rData.ID = tfData.Modules[i].Resources[k].Primary[p].ID
				}
			}
			rList = append(rList, rData)
		}
	}

	log.Printf("Total (%d) resource in (%s).\n", len(rList), terraformerStateFile)
	return rList
}

// ReadTerraformStateFile ..
// TF 0.13+ compatible
func ReadTerraformStateFile(terraformStateFile, repoType string) map[string]interface{} {
	rIDs := make(map[string]interface{})
	tfData := TerraformSate{}

	tfFile, err := ioutil.ReadFile(terraformStateFile)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal([]byte(tfFile), &tfData)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(tfData.Resources); i++ {
		rData := Resource{}
		var key string
		//Don't process the mode type with 'data' value
		if tfData.Resources[i].Mode == "data" {
			continue
		}

		rData.ResourceName = tfData.Resources[i].ResourceName
		rData.ResourceType = tfData.Resources[i].ResourceType
		for k := 0; k < len(tfData.Resources[i].Instances); k++ {
			rData.ID = tfData.Resources[i].Instances[k].Attributes.ID
			if tfData.Resources[i].Instances[k].DependsOn != nil {
				rData.DependsOn = tfData.Resources[i].Instances[k].DependsOn
			}

			if repoType == "discovery" {
				key = rData.ResourceType + "." + rData.ResourceName
			} else {
				key = rData.ResourceType + "." + rData.ID
			}
			rData.ResourceIndex = i
			rIDs[key] = rData
		}
	}

	log.Printf("Total (%d) resource in (%s).\n", len(rIDs), terraformStateFile)
	return rIDs
}

// DiscoveryImport ..
func DiscoveryImport(configName, services, tags, randomID, discoveryDir string) error {
	log.Printf("# let's import the resources (%s) 2/6:\n", services)

	// Import the terraform resources & state files.
	err := TerraformerImport(discoveryDir, services, tags, configName, &planTimeOut, randomID)
	if err != nil {
		return err
	}

	log.Println("# Writing HCL Done!")
	log.Println("# Writing TFState Done!")

	//Check terraform version compatible
	log.Println("# now, we can do some infra as code ! First, update the IBM Terraform provider to support TF 0.13 [3/6]:")
	err = UpdateProviderFile(discoveryDir, randomID, &planTimeOut)
	if err != nil {
		return err
	}

	//Run terraform init commnd
	log.Println("# we need to init our Terraform project [4/6]:")
	err = TerraformInit(discoveryDir, "", &planTimeOut, randomID)
	if err != nil {
		return err
	}

	//Run terraform refresh commnd on the generated state file
	log.Println("# and finally compare what we imported with what we currently have [5/6]:")
	err = TerraformRefresh(discoveryDir, "", &planTimeOut, randomID)
	if err != nil {
		return err
	}

	return nil
}

// UpdateProviderFile ..
func UpdateProviderFile(discoveryDir, randomID string, timeout *time.Duration) error {
	providerTF := discoveryDir + "/provider.tf"
	input, err := ioutil.ReadFile(providerTF)
	if err != nil {
		return err
	}

	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		if strings.Contains(line, "version") {
			lines[i] = "source = \"IBM-Cloud/ibm\""
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(providerTF, []byte(output), 0644)
	if err != nil {
		return err
	}

	//Replace provider path in state file
	err = TerraformReplaceProvider(discoveryDir, randomID, &planTimeOut)
	if err != nil {
		return err
	}
	return nil
}

// MergeStateFile ..
func MergeStateFile(configRepoMap, discoveryRepoMap map[string]interface{}, src, dest, configDir, scenario, randomID string, timeout *time.Duration) error {
	var mergeResourceList []string

	//Read discovery state file
	content, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	statefilecontent := string(content)

	//Loop through each discovery repo resource with config repo resource
	for _, dResource := range discoveryRepoMap {
		//Discovery resource
		discovery_resource := dResource.(Resource).ResourceType + "." + dResource.(Resource).ID

		//Check discovery resource exist in config repo.
		//If resource not exist, Move the discovery resource to config repo
		if configRepoMap[discovery_resource] == nil {
			discovery_resource := dResource.(Resource).ResourceType + "." + dResource.(Resource).ResourceName
			mergeResourceList = append(mergeResourceList, discovery_resource)
		} else {
			//Resource allready exist in config repo
			continue
		}

		//Check discovery resource has got depends_on attribute
		//If depends_on attribute exist in discovery resource, Get the depends_on resource name from config repo & update in discovery state file.
		if dResource.(Resource).DependsOn != nil {
			var dependsOn []string
			for i, d := range dResource.(Resource).DependsOn {
				configParentResource := discoveryRepoMap[d].(Resource).ResourceType + "." + discoveryRepoMap[d].(Resource).ID

				//Get parent resource from config repo
				if configRepoMap[configParentResource] != nil {
					//Get depends_on resource name from config repo to update in discovery state file
					configParentResource = configRepoMap[configParentResource].(Resource).ResourceType + "." + configRepoMap[configParentResource].(Resource).ResourceName
					dependsOn = append(dependsOn, configParentResource)

					//Update depends_on parameter in discovery state file content
					statefilecontent, err = sjson.Set(statefilecontent, "resources."+strconv.Itoa(dResource.(Resource).ResourceIndex)+".instances.0.dependencies."+strconv.Itoa(i), configParentResource)
					if err != nil {
						return err
					}
				}
			}

			//Copy the state file content changes to discovery repo state file
			if len(dependsOn) > 0 {
				err = ioutil.WriteFile(src, []byte(statefilecontent), 0644)
				if err != nil {
					return err
				}
			}
		}
	}

	//Move resource from discovery repo to config repo state file
	if len(mergeResourceList) > 0 {
		for _, resource := range mergeResourceList {
			err = TerraformMoveResource(configDir, src, dest, resource, "", &planTimeOut, randomID)
			if err != nil {
				return err
			}
		}
		log.Printf("\n\n# Discovery service successfuly moved (%v) resources from (%s) to (%s).", len(mergeResourceList), src, dest)
	} else {
		log.Printf("\n\n# Discovery service didn't find any resource to move from (%s) to (%s).", src, dest)
	}

	return nil
}
