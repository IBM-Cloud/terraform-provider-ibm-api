// +build darwin

package main

import (
	"log"
	"os"
)

const tfrName = "terraformer"
const tfName = "terraform"

var goPath = os.Getenv("GOPATH")

// separate out command to install terraform and terraformer
// Add separate compile time files for each os - 3
// take versions (default to latest) and provide output as installed path
// ask user to add to the path

func downloadAndInitialize(tfrVersion, tfVersion, installPath, providerVersio string) error {
	log.Printf("versions: terraformer %s and terraform %s\n", tfrVersion, tfVersion)

	log.Println("TODO: Download terraformer and terraform here, if not available")

	if installPath != "" {
		if _, err := os.Stat(installPath); os.IsExist(err) {

		} else {

		}
	} else if goPath != "" {
		// Install in go path
		installPath = goPath
	} else {
		// Download into current folder and ask user to put it in path somewhere
		var err error
		installPath, err = os.Getwd()
		if err != nil {
			log.Println("Error with getting current folder, when downloading into current folder")
			return err
		}
	}

	// todo: @srikar - check existence of terraform
	if false {
		err := downloadTerraform(tfVersion)
		if err != nil {
			log.Println("Error with downloading terraform")
			return err
		}
	}

	// todo: @srikar - check existence of terraformer
	if false {
		err := downloadTar(tfrVersion, tfrGHRepo, "", "", true, true)
		if err != nil {
			log.Println("Error with downloading terraformer")
			return err
		}
	}

	//  // todo: @srikar -  where to place provider
	// todo: @srikar - check existence of terraform provider
	if false {
		err := downloadTar(tfProviderGHRepo, tfrGHRepo, "", "", true, true)
		if err != nil {
			log.Println("Error with downloading terraform provider")
			return err
		}
		// todo: @srikar - Place this in right place
	}

	return nil
}

func downloadTerraform(version string) error {
	return nil
}
