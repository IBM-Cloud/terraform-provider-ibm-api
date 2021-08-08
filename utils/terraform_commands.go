package utils

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"time"
)

//TerraformInit ...
func TerraformInit(configDir string, scenario string, timeout *time.Duration, randomID string) error {

	return Run("terraform", []string{"init"}, configDir, scenario, timeout, randomID)
}

//TerraformApply ...
func TerraformApply(configDir, stateDir string, scenario string, timeout *time.Duration, randomID string) error {
	return Run("terraform", []string{"apply", fmt.Sprintf("-state=%s", stateDir+"/"+scenario+".tfstate"), "-auto-approve"}, configDir, scenario, timeout, randomID)
}

//TerraformPlan ...
func TerraformPlan(configDir string, scenario string, timeout *time.Duration, randomID string) error {
	return Run("terraform", []string{"plan"}, configDir, scenario, timeout, randomID)
}

//TerraformDestroy ...
func TerraformDestroy(configDir, stateDir string, scenario string, timeout *time.Duration, randomID string) error {

	return Run("terraform", []string{"destroy", "-force", fmt.Sprintf("-state=%s", stateDir+"/"+scenario+".tfstate")}, configDir, scenario, timeout, randomID)
}

//TerraformShow ...
func TerraformShow(configDir, stateDir string, scenario string, timeout *time.Duration, randomID string) error {

	return Run("terraform", []string{"show", fmt.Sprintf("%s", stateDir+"/"+scenario+".tfstate")}, configDir, scenario, timeout, randomID)
}

func Run(cmdName string, args []string, configDir string, scenario string, timeout *time.Duration, randomID string) error {
	cmd := exec.Command(cmdName, args...)
	if timeout != nil {
		ctx, cancel := context.WithTimeout(context.Background(), *timeout)
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
	fmt.Println("Starting command", cmd.Path, cmd.Args)
	err = cmd.Start()
	if err != nil {
		return err
	}

	//Wait for command to finish
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
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

	if _, err := os.Stat(stderrPath); err == nil {
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
