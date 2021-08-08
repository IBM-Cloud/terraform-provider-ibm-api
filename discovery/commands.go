package discovery

import (
	"fmt"
	"time"

	"github.com/IBM-Cloud/terraform-provider-ibm-api/utils"
)

//TerraformerImport ...
func TerraformerImport(configDir string, opts []string, scenario string, timeout *time.Duration, randomID string) error {
	return utils.Run("terraformer", append([]string{"import", "ibm", "--compact", fmt.Sprintf("-p=%s", configDir)}, opts...), configDir, scenario, timeout, randomID)
}

//TerraformMoveResource ...
func TerraformMoveResource(configDir, srcStateFile, destStateFile, resourceName, scenario string, timeout *time.Duration, randomID string) error {

	return utils.Run("terraform", []string{"state", "mv", fmt.Sprintf("-state=%s", srcStateFile), fmt.Sprintf("-state-out=%s", destStateFile), resourceName, resourceName}, configDir, scenario, timeout, randomID)
}

//TerraformReplaceProvider ..
func TerraformReplaceProvider(configDir, randomID string, timeout *time.Duration) error {
	//terraform state
	return utils.Run("terraform", []string{"state", "replace-provider", "-auto-approve", "registry.terraform.io/-/ibm", "registry.terraform.io/ibm-cloud/ibm"}, configDir, "", timeout, randomID)
}

//TerraformRefresh ...
func TerraformRefresh(configDir string, scenario string, timeout *time.Duration, randomID string) error {
	return utils.Run("terraform", []string{"refresh"}, configDir, scenario, timeout, randomID)
}
