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
	"time"
)

// TerraformInit ...
func TerraformInit(execDir string, timeout *time.Duration, randomID string) error {

	return Run("terraform", []string{"init"}, execDir, timeout, randomID)
}

// TerraformApply ...
func TerraformApply(execDir, stateDir string, stateFileName string, timeout *time.Duration, randomID string) error {
	return Run("terraform", []string{"apply", fmt.Sprintf("-state=%s", stateDir+pathSep+stateFileName+".tfstate"), "-auto-approve"}, execDir, timeout, randomID)
}

// TerraformPlan ...
func TerraformPlan(execDir string, timeout *time.Duration, randomID string) error {
	return Run("terraform", []string{"plan"}, execDir, timeout, randomID)
}

// TerraformDestroy ...
func TerraformDestroy(execDir, stateDir string, stateFileName string, timeout *time.Duration, randomID string) error {

	return Run("terraform", []string{"destroy", "-force", fmt.Sprintf("-state=%s", stateDir+pathSep+stateFileName+".tfstate")}, execDir, timeout, randomID)
}

// TerraformShow ...
func TerraformShow(execDir, stateDir string, stateFileName string, timeout *time.Duration, randomID string) error {

	return Run("terraform", []string{"show", stateDir + pathSep + stateFileName + ".tfstate"}, execDir, timeout, randomID)
}

func Run(cmdName string, args []string, execDir string, timeout *time.Duration, randomID string) error {
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

	cmd.Dir = execDir

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

func getLogFiles(logDir, fileName string) (stdoutFile, stderrFile *os.File, err error) {
	stdoutPath := path.Join(logDir, fileName+".out")
	stderrPath := path.Join(logDir, fileName+".err")

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
