package discovery

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM-Cloud/terraform-provider-ibm-api/utils"
)

//TerraformerImport ...
func TerraformerImport(configDir string, opts []string, timeout time.Duration, randomID string) error {
	return utils.Run(context.Background(), "terraformer", append([]string{"import", "ibm", fmt.Sprintf("-p=%s", configDir)}, opts...), configDir, timeout, randomID)
}

//TerraformMoveResource ...
func TerraformMoveResource(configDir, srcStateFile, destStateFile, resourceName string, timeout time.Duration, randomID string) error {

	return utils.Run(context.Background(), "terraform", []string{"state", "mv", fmt.Sprintf("-state=%s", srcStateFile), fmt.Sprintf("-state-out=%s", destStateFile), resourceName, resourceName}, configDir, timeout, randomID)
}

//TerraformReplaceProvider ..
func TerraformReplaceProvider(configDir, randomID string, timeout time.Duration) error {
	//terraform state
	return utils.Run(context.Background(), "terraform", []string{"state", "replace-provider", "-auto-approve", "registry.terraform.io/-/ibm", "registry.terraform.io/ibm-cloud/ibm"}, configDir, timeout, randomID)
}

//TerraformRefresh ...
func TerraformRefresh(configDir string, timeout time.Duration, randomID string) error {
	return utils.Run(context.Background(), "terraform", []string{"refresh"}, configDir, timeout, randomID)
}
