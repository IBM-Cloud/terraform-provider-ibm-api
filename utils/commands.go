package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
)

var stdouterr []byte

//It will clone the git repo which contains the configuration file.
func cloneRepo(msg ConfigRequest) ([]byte, string, error) {
	gitURL := msg.GitURL
	urlPath, err := url.Parse(msg.GitURL)
	if err != nil {
		return nil, "", err
	}
	baseName := filepath.Base(urlPath.Path)
	extName := filepath.Ext(urlPath.Path)
	p := baseName[:len(baseName)-len(extName)]
	if _, err := os.Stat(currentDir + "/" + p); err == nil {
		stdouterr, err = pullRepo(p)

	} else {
		cmd := exec.Command("git", "clone", gitURL)
		fmt.Println(cmd.Args)
		cmd.Dir = currentDir
		stdouterr, err = cmd.CombinedOutput()
		if err != nil {
			return nil, "", err
		}
	}
	path := currentDir + "/" + p + "/terraform.tfvars"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		createFile(msg, path)
	} else {
		err = os.Remove(path)
		createFile(msg, path)
	}

	return stdouterr, p, err
}

//It will create a vars file
func createFile(msg ConfigRequest, path string) {
	// detect if file exists

	_, err := os.Stat(path)

	// create file if not exists
	if os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return
		}
		defer file.Close()
	}

	writeFile(path, msg)
}

func createFiles(path string) {
	// check if file exists
	var _, err = os.Stat(path)

	// create file if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		if isError(err) {
			return
		}
		defer file.Close()
	}

	fmt.Println("File Created Successfully", path)
}

func pullRepo(repoName string) ([]byte, error) {
	cmd := exec.Command("git", "pull")
	fmt.Println(cmd.Args)
	cmd.Dir = currentDir + "/" + repoName
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return stdoutStderr, err
}

func removeRepo(path, repoName string) error {
	removePath := filepath.Join(path, repoName)
	err := os.RemoveAll(removePath)
	return err
}

func writeFile(path string, msg ConfigRequest) {
	// open file using READ & WRITE permission
	var file, err = os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	variables := msg.VariableStore

	if variables != nil {
		for _, v := range *variables {
			_, err = file.WriteString(v.Name + " = \"" + v.Value + "\" \n")
		}
	}

	// save changes
	err = file.Sync()
	if err != nil {
		return
	}
}

func deleteFile(path string) {
	// delete file
	var err = os.Remove(path)
	if isError(err) {
		return
	}

	log.Println("File Deleted")
}

func removeDir(path string) {
	// Remove all the directories and files
	// Using RemoveAll() function
	err := os.RemoveAll(path)
	if err != nil {
		log.Fatal(err)
	}
}

// Copy ..
func Copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func isError(err error) bool {
	if err != nil {
		log.Println(err.Error())
	}

	return (err != nil)
}

// ReadTerraformerStateFile ..
func ReadTerraformerStateFile(terraformerStateFile string) ResourceList {
	var rList ResourceList
	file, _ := ioutil.ReadFile(terraformerStateFile)
	tfData := TerraformSate{}
	_ = json.Unmarshal([]byte(file), &tfData)
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
	fmt.Println("Terraformer state ::", len(rList))
	return rList
}

// ReadTerraformStateFile ..
func ReadTerraformStateFile(terraformStateFile string) ResourceList {
	var rList ResourceList
	file, _ := ioutil.ReadFile(terraformStateFile)
	tfData := TerraformSate{}
	_ = json.Unmarshal([]byte(file), &tfData)
	for i := 0; i < len(tfData.Resources); i++ {
		rData := Resource{}
		rData.ResourceName = tfData.Resources[i].ResourceName
		rData.ResourceType = tfData.Resources[i].ResourceType
		if tfData.Resources[i].Mode != "data" {
			for k := 0; k < len(tfData.Resources[i].Instances); k++ {
				rData.ID = tfData.Resources[i].Instances[k].Attributes.ID
			}
			rList = append(rList, rData)
		}
	}
	fmt.Println("Terraform state ::", len(rList))
	return rList
}
