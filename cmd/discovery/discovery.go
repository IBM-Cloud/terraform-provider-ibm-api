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
	app.Usage = "Lets you create state file and TF Config from Resources in your cloud account. " +
		"For the green field and brown field imports of config and statefile, " +
		"and all terraformer related"
	// app.Writer = ui.Writer()
	// app.ErrWriter = ui.Writer()

	// we create our commands
	app.Commands = []cli.Command{
		{
			Category:    "discovery",
			Name:        "version",
			Description: "Version",
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
			Name:     "dep",
			Aliases:  []string{"dependency, dependencies, download"},
			Usage: fmt.Sprint(
				"discovery",
				" dep",
				" [--terraformer_version terraformer_version]",
				" [--terraform_version terraform_version]",
			),
			Description: "Installs terraformer " +
				"and terraform if binaries are not found. Version matters only if binary is not already " +
				"present.",
			Flags: []cli.Flag{
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
				cli.StringFlag{
					Name:  "path",
					Usage: "Install the binaries here, if not given defaults to " + defaultPath,
				},
			},

			Action: func(c *cli.Context) error {

				tfrVersion := c.String("terraformer_version")
				tfVersion := c.String("terraform_version")
				installPath := c.String("path")

				err := downloadAndInitialize(tfrVersion, tfVersion, installPath)
				if err != nil {
					log.Println("ERROR: Couldn't check and download tf and tfer binaries")
					return err
				}

				// downloaded here, add these to your path
				return nil
			},
			OnUsageError: func(ctx *cli.Context, err error, isSub bool) error {
				return cli.ShowCommandHelp(ctx, ctx.Command.Name)
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
				" [--config_dir CONFIG_DIR]",
			),
			Description: "Clone and create a local configuration in an empty repo and run terraform " +
				"init. Clones in to a directory (printed with name config_name) " +
				"If git_url is not passed, terraformer and terraform will be " +
				"installed and terraform init is done in the config_dir. " +
				"Set TF_LOG like you set for terraform for debug logs in your env ",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "git_url",
					Usage: "The git url to get the configuration from. If empty, config_dir should have tf files.",
				},
				cli.StringFlag{
					Name: "config_dir",
					Usage: "If git_url is passed, Must be an empty existing folder. A folder to operate in. " +
						"If git_url is not passed, this folder should have tf files already. In this case, " +
						"empty means current folder. Can be used to download terraformer and terraform.",
					Value: "./",
				},
			},

			Action: func(c *cli.Context) error {

				gitURL := c.String("git_url")
				confDir := c.String("config_dir")

				log.Println("git_url", gitURL)
				log.Println("config dir", confDir)

				var repoName string
				var err error
				if gitURL == "" {
					log.Println("EMPTY GIT URL: No git_url given, skipping to tf init")
				} else {

					if err := createDirs(confDir, false); err != nil {
						return err
					}
					log.Println("Will clone git repo", gitURL)

					_, repoName, err = utils.CloneRepo(utils.ConfigRequest{
						GitURL: gitURL,
					})
					if err != nil {
						log.Println("Eror Cloning repo..")
						log.Printf("err : %v\n", err)
						return err
					}
					log.Println("\n config name: ", repoName)
				}

				b := make([]byte, 10)
				rand.Read(b)
				randomID := fmt.Sprintf("%x", b)

				// todo repoName passed as scenario
				err = utils.TerraformInit(path.Join(confDir, repoName), repoName, &planTimeOut, randomID)
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
			Description: "Import TF config for resources in your ibm cloud account. " +
				"Import all the resources for this service. Imports config and statefile. " +
				"If a statefile is already present, merging will be done. ",
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
					Usage: "Folder inside config_dir, your config dir, where config was generated " +
						"config_name outputted in config command if you've cloned. Leave empty if " +
						"configured a local directory.",
					Value: "",
				},
			},

			Action: func(c *cli.Context) error {

				if !c.IsSet("services") {
					return fmt.Errorf("services flag not set")
				}

				confDir := c.String("config_dir")
				services := c.String("services")
				repoName := c.String("repo_name") // todo: @srikar - change to repo_name from conf_name evrywhr
				command := c.String("command")
				tags := c.String("tags")

				opts := []string{}
				func() {
					if services != "" {
						opts = append(opts, "--resources="+services)
					}
					if tags != "" {
						fmt.Println(tags)
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

				if err := createDirs(confDir, true); err != nil {
					return err
				}

				b := make([]byte, 10)
				rand.Read(b)
				randomID := fmt.Sprintf("%x", b)

				log.Println("Backend random id created: Intermediate state here", randomID)

				//Clean up discovery directory
				discoveryDir := confDir + "/" + "discovery"
				err := utils.RemoveDir(discoveryDir + "/*")
				if err != nil {
					// todo: @srikar - log error
					return err
				}

				if command == "default" {
					if repoName != "discovery" {
						discoveryDir = confDir + "/" + repoName
					}
					err = utils.DiscoveryImport(repoName, randomID, discoveryDir, opts)
					if err != nil {
						// todo: @srikar - log error
						return err
					}
				} else if command == "merge" {

					// Import the terraform resources & state files.
					// err := utils.TerraformerImport(confDir, services, repoName, &planTimeOut, randomID)
					err = utils.DiscoveryImport(repoName, randomID, discoveryDir, opts)
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
					repoDir := confDir + "/" + repoName
					//Backup repo TF file.
					terraformStateFile := repoDir + "/terraform.tfstate"
					err = utils.Copy(terraformStateFile, repoDir+"/terraform.tfstate_backup")
					if err != nil {
						log.Println("Error with copying file")
						return err
					}

					if _, err := os.Stat(terraformStateFile); os.IsNotExist(err) {
						log.Printf("No merging needed bcz statefile doesn't already exist at %s\n", terraformStateFile)
						log.Println("Done. Exiting")
						return nil
					}

					terraformObj := utils.ReadTerraformStateFile(terraformStateFile, "")

					//Read state file from discovery repo directory
					// terraformerStateFile := confDir + "/generated" + "/ibm/" + srv + "/terraform.tfstate"
					terraformerSateFile := discoveryDir + "/terraform.tfstate"
					terraformerObj := utils.ReadTerraformStateFile(terraformerSateFile, "discovery")

					// comparing state files
					if cmp.Equal(terraformObj, terraformerObj, cmpopts.IgnoreFields(utils.Resource{}, "ResourceName")) {
						log.Println("# Config repo configuration/state is equal !!")
					} else {
						log.Println("# Config repo configuration/state is not equal !!")
						// utils.MergeStateFile(terraformObj, terraformerObj, terraformerSateFile, terraformStateFile,"", "", randomID, &planTimeOut)
						err = utils.MergeStateFile(terraformObj, terraformerObj, terraformerSateFile, terraformStateFile, confDir, "", randomID, &planTimeOut)
						if err != nil {
							// todo: @srikar - handle error
							return err
						}
					}

					log.Println("Backend action: file state here finally", terraformStateFile)
				}

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
