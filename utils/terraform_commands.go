package utils

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"time"
)

// TerraformInit ...
func TerraformInit(configDir string, scenario string, timeout *time.Duration, randomID string) error {

	return run("terraform", []string{"init"}, configDir, scenario, timeout, randomID)
}

// TerraformApply ...
func TerraformApply(configDir, stateDir string, scenario string, timeout *time.Duration, randomID string) error {
	return run("terraform", []string{"apply", fmt.Sprintf("-state=%s", stateDir+"/"+scenario+".tfstate"), "-auto-approve"}, configDir, scenario, timeout, randomID)
}

// TerraformPlan ...
func TerraformPlan(configDir string, scenario string, timeout *time.Duration, randomID string) error {
	return run("terraform", []string{"plan"}, configDir, scenario, timeout, randomID)
}

// TerraformRefresh ...
func TerraformRefresh(configDir string, scenario string, timeout *time.Duration, randomID string) error {
	return run("terraform", []string{"refresh"}, configDir, scenario, timeout, randomID)
}

// TerraformDestroy ...
func TerraformDestroy(configDir, stateDir string, scenario string, timeout *time.Duration, randomID string) error {

	return run("terraform", []string{"destroy", "-force", fmt.Sprintf("-state=%s", stateDir+"/"+scenario+".tfstate")}, configDir, scenario, timeout, randomID)
}

// TerraformShow ...
func TerraformShow(configDir, stateDir string, scenario string, timeout *time.Duration, randomID string) error {

	return run("terraform", []string{"show", stateDir + "/" + scenario + ".tfstate"}, configDir, scenario, timeout, randomID)
}

// TerraformerImport ...
func TerraformerImport(configDir string, opts []string, scenario string, timeout *time.Duration, randomID string) error {
	return run("terraformer", append([]string{"import", "ibm", "--compact", fmt.Sprintf("-p=%s", configDir)}, opts...), configDir, scenario, timeout, randomID)
}

// TerraformMoveResource ...
func TerraformMoveResource(configDir, srcStateFile, destStateFile, resourceName, scenario string, timeout *time.Duration, randomID string) error {

	return run("terraform", []string{"state", "mv", fmt.Sprintf("-state=%s", srcStateFile), fmt.Sprintf("-state-out=%s", destStateFile), resourceName, resourceName}, configDir, scenario, timeout, randomID)
}

// TerraformReplaceProvider ..
func TerraformReplaceProvider(configDir, randomID string, timeout *time.Duration) error {
	//terraform state
	return run("terraform", []string{"state", "replace-provider", "-auto-approve", "registry.terraform.io/-/ibm", "registry.terraform.io/ibm-cloud/ibm"}, configDir, "", timeout, randomID)
}

func run(cmdName string, args []string, configDir string, scenario string, timeout *time.Duration, randomID string) error {
	cmd := exec.Command(cmdName, args...)
	if timeout != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		cmd = exec.CommandContext(ctx, cmdName, args...)
		defer cancel()
	}

	stdoutFile, stderrFile, err := getLogFiles(logDir, randomID)
	if err != nil {
		return err
	}
	defer stdoutFile.Close()
	defer stderrFile.Close()

	cmd.Dir = configDir

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	//Write the stdout to log file
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fmt.Fprintln(stdoutFile, scanner.Text())
		}
	}()

	//Write the stderr to log file
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Fprintln(stderrFile, scanner.Text())

		}
	}()

	//Start the command
	log.Println("Starting command", cmd.Path, cmd.Args)
	err = cmd.Start()
	if err != nil {
		return err
	}

	//Wait for command to finish
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return err
}

func getLogFiles(logDir, scenario string) (stdoutFile, stderrFile *os.File, err error) {
	stdoutPath := path.Join(logDir, scenario+".out")
	stderrPath := path.Join(logDir, scenario+".err")

	if _, err = os.Stat(stdoutPath); err == nil {
		stdoutFile, err = os.OpenFile(stdoutPath, os.O_APPEND|os.O_WRONLY, 0600)
	} else {
		stdoutFile, err = os.Create(stdoutPath)
	}
	if err != nil {
		return
	}

	if _, err = os.Stat(stderrPath); err == nil {
		stderrFile, err = os.OpenFile(stderrPath, os.O_APPEND|os.O_WRONLY, 0600)
	} else {
		stderrFile, err = os.Create(stderrPath)
	}
	return
}

func readLogFile(logID string) (stdout, stderr string, err error) {
	stdoutPath := path.Join(logDir, logID+".out")
	stderrPath := path.Join(logDir, logID+".err")

	outFile, err := ioutil.ReadFile(stdoutPath)
	if err != nil {
		return
	}
	errFile, err := ioutil.ReadFile(stderrPath)

	if err != nil {
		return
	}

	return string(outFile), string(errFile), nil
}

func MergeResources(path string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	createFiles(terraformerfWrapperDir + "/main.tf")

	rg, _ := regexp.Compile("^(output|main|terraform|provider)")
	for _, f := range files {
		if !rg.MatchString(f.Name()) {
			data, err := ioutil.ReadFile(path + "/" + f.Name())
			if err != nil {
				fmt.Println("File reading error :", err)
				return err
			}

			file, err := os.OpenFile(terraformerfWrapperDir+"/main.tf", os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				log.Fatalf("failed opening file: %s", err)
			}
			defer file.Close()

			_, err = file.WriteString(string(data))
			if err != nil {
				log.Fatalf("failed writing to file: %s", err)
			}
		}
	}

	//Delete backup files from terraform_wrapper
	fileList, err := filepath.Glob(terraformerfWrapperDir + "/terraform.tfstate.*")
	if err != nil {
		log.Fatalf("failed removing file: %s", err)
	}
	for _, f := range fileList {
		if err := os.Remove(f); err != nil {
			log.Fatalf("failed removing file: %s", err)
		}
	}

	return nil
}
