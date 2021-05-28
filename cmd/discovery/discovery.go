package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/IBM-Cloud/terraform-provider-ibm-api/utils"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/urfave/cli"
)

var planTimeOut = 60 * time.Minute

// todo - can users set this directly
// if msg.LOGLEVEL != "" {
// 	os.Setenv("TF_LOG", msg.LOGLEVEL)
// } user can set this directly

func main() {

	// ui := terminal.NewStdUI()
	app := cli.NewApp()
	app.Name = "discovery"
	app.HelpName = "IBM Cloud Discovery CLI"
	app.Usage = `Lets you create state file and TF Config from Resources in your cloud account. 
	For the green field and brown field imports of config and statefile, and all terraformer related`
	// app.Writer = ui.Writer()
	// app.ErrWriter = ui.Writer()

	// we create our commands
	app.Commands = []cli.Command{
		{
			Category:    "discovery",
			Name:        "version",
			Description: `Version`,
			Usage: fmt.Sprint(
				"discovery",
				" version",
			),
			Action: func(ctx *cli.Context) error {
				fmt.Println("0.1.0")
				return nil
			},
			OnUsageError: func(ctx *cli.Context, err error, isSub bool) error {
				return cli.ShowCommandHelp(ctx, ctx.Args().First())
			},
		},
		{
			Category: "discovery",
			Name:     "config",
			Aliases:  []string{"configure"},
			Usage: fmt.Sprint(
				"discovery",
				" config",
				" [--git_url GIT_URL]",
				" [--service SERVICE_TO_IMPORT]",
				" [--config_dir CONFIG_DIR]",
				" [--terraformer_version terraformer_version]",
				" [--terraform_version terraform_version]",
				// individual resources to be specified ?
			),
			Description: `Clone and create a local configuration in an empty repo and run terraform init.
					Clones in to a directory (printed with name config_name)
					Installs terraformer and terraform if binaries are not found. Version matters only if 
					binary is not already present. If git_url is not passed, terraformer and terraform will be
					installed and terraform init is done in the config_dir.
					Set TF_LOG like you set for terraform for debug logs in your env `,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "git_url",
					Usage: "The git url to get the configuration from. If empty, config_dir should have tf files.",
				},
				cli.StringFlag{ // todo is service needed here
					Name:  "services",
					Usage: "The IBM service to import the resources from",
				},
				cli.StringFlag{
					Name: "config_dir",
					Usage: `If git_url is passed, Must be an empty existing folder. A folder to operate in. 
							If git_url is not passed, this folder should have tf files already. In this case, 
							empty means current folder. Can be used to download terraformer and terraform.`,
					Value: "./",
				},
				cli.StringFlag{
					Name:  "terraformer_version",
					Value: "latest",
					Usage: "If terraformer binary is not found, this version will be installed. Defaults to latest",
				},
				cli.StringFlag{
					Name:  "terraform_version",
					Value: "latest",
					Usage: "If terraform binary is not found, this version will be installed. Defaults to latest",
				},
			},

			Action: func(c *cli.Context) error {

				tfrVersion := c.String("terraformer_version")
				tfVersion := c.String("terraform_version")

				err := downloadAndInitialize(tfrVersion, tfVersion)
				if err != nil {
					log.Println("ERROR: Couldn't check and download tf and tfer binaries")
					return err
				}

				gitURL := c.String("git_url")
				services := c.String("services")
				confDir := c.String("config_dir")

				log.Println("git_url", gitURL)
				log.Println("config dir", confDir)
				log.Println("services passed", services)

				var configName string
				if gitURL == "" {
					log.Println("EMPTY GIT URL: No git_url given, skipping to tf init")
				} else {

					if err := createDirs(confDir, false); err != nil {
						return err
					}
					log.Println("Will clone git repo", gitURL)

					_, configName, err = utils.CloneRepo(utils.ConfigRequest{
						GitURL:  gitURL,
						Service: services,
					})
					if err != nil {
						log.Println("Eror Cloning repo..")
						log.Printf("err : %v\n", err)
						return err
					}
					log.Println("\n config name: ", configName)
				}

				b := make([]byte, 10)
				rand.Read(b)
				randomID := fmt.Sprintf("%x", b)

				// todo configName passed as scenario
				err = utils.TerraformInit(path.Join(confDir, configName), configName, &planTimeOut, randomID)
				if err != nil {
					log.Println("TF INIT ERROR:", err.Error())
					return err
				}
				return nil
			},
			OnUsageError: func(ctx *cli.Context, err error, isSub bool) error {
				return cli.ShowCommandHelp(ctx, ctx.Command.Name)
			},
		},
		{
			Category: "discovery",
			Name:     "import",
			Usage: fmt.Sprint(
				"discovery",
				" import",
				" --services SERVICE_TO_IMPORT", // ibm_is_instance
				" [--config_dir CONFIG_DIR]",
				" [--config_name CONFIG_NAME]",
				// todo individual resources to be specified ?
			),
			Description: `Import TF config for resources in your ibm cloud account. 
					Import all the resources for this service. Imports config and statefile. 
					If a statefile is already present, merging will be done.`,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "services",
					Usage: "The IBM service(s) to import the resources from. Comma separated",
				},
				cli.StringFlag{
					Name:  "config_dir",
					Usage: `Empty means current folder.`,
					Value: "./",
				},
				cli.StringFlag{
					Name: "config_name",
					Usage: `Folder inside config_dir, your config dir, where config was generated
							config_name outputted in config command if you've cloned. Leave empty if 
							configured a local directory.`,
					Value: "",
				},
			},

			Action: func(c *cli.Context) error {

				if !c.IsSet("services") {
					return fmt.Errorf("service flag not set")
				}

				// Read Query Parameter
				confDir := c.String("config_dir")
				services := c.String("services")
				configName := c.String("config_name")

				if err := createDirs(confDir, true); err != nil {
					return err
				}

				b := make([]byte, 10)
				rand.Read(b)
				randomID := fmt.Sprintf("%x", b)

				log.Println("Backend random id created: Intermediate state here", randomID)

				// Import the terraform resources & state files.
				err := utils.TerraformerImport(confDir, services, configName, &planTimeOut, randomID)
				if err != nil {
					log.Println("Error with importing", err)
					return err
				}

				if _, err := os.Stat(confDir + "/generated" + "/ibm"); os.IsNotExist(err) {
					log.Println("Import not successful")
					return nil
				} else {
					log.Printf("Import successful. Imported into %s\n", confDir+"/generated"+"/ibm")
				}

				//Merge state files and templates in services
				repoDir := confDir + "/" + configName
				terraformStateFile := repoDir + "/terraform.tfstate"
				if _, err := os.Stat(terraformStateFile); os.IsNotExist(err) {
					log.Printf("No merging needed bcz statefile doesn't already exist at %s\n", terraformStateFile)
					log.Println("Done. Exiting")
					return nil
				}
				//Backup repo TF file.
				err = utils.Copy(terraformStateFile, repoDir+"/terraform.tfstate_backup")
				if err != nil {
					log.Println("Error with copying file")
					return err
				}

				//Read resources from terraform state file
				terraformObj := utils.ReadTerraformStateFile(terraformStateFile)
				service := strings.Split(services, ",")
				if len(service) > 0 {
					for _, srv := range service {
						//Read resources from terraformer state file
						terraformerSateFile := confDir + "/generated" + "/ibm/" + srv + "/terraform.tfstate"
						terraformerObj := utils.ReadTerraformerStateFile(terraformerSateFile)

						// comparing state files
						if cmp.Equal(terraformObj, terraformerObj, cmpopts.IgnoreFields(utils.Resource{}, "ResourceName")) {
							log.Println("State is equal..")
						} else {
							log.Println("State is not equal..")
							utils.CompareStateFile(terraformObj, terraformerObj, terraformerSateFile, terraformStateFile, "", "", randomID, &planTimeOut)
						}
					}
				}

				log.Println("Backend action: file state here finally", terraformStateFile)

				log.Println("Successful import")

				return nil
			},
			OnUsageError: func(ctx *cli.Context, err error, isSub bool) error {
				return cli.ShowCommandHelp(ctx, ctx.Command.Name)
			},
		},
	}

	// start our application
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func createDirs(confDir string, imp bool) (err error) {
	defer func() {
		if err != nil {
			log.Println("ERRROR in creating directories")
		}
	}()
	if _, err = os.Stat(confDir); os.IsNotExist(err) {
		log.Println("ERROR: Folder doesn't exist", confDir)
		return err
	}

	if imp {
		// if _, err = os.Stat(confDir + "/log/"); os.IsNotExist(err) {
		// 	err = os.MkdirAll(confDir+"/log/", os.ModePerm)
		// 	if err != nil {
		// 		return err
		// 	}
		// }
		if _, err = os.Stat(confDir + "/state"); os.IsNotExist(err) {
			err = os.MkdirAll(confDir+"/state", os.ModePerm)
			if err != nil {
				return err
			}
		}

		if _, err := os.Stat(confDir + "/terraformer_wrapper"); os.IsNotExist(err) {
			err := os.MkdirAll(confDir+"/terraformer_wrapper", os.ModePerm)
			if err != nil {
				return err
			}
		}
		utils.SetGlobalDirs(
			confDir,
			"", //confDir+"/log/",
			confDir+"/state",
			confDir+"/terraformer_wrapper",
		)
	} else {
		utils.SetGlobalDirs(confDir, "", "", "")
	}
	return nil
}
