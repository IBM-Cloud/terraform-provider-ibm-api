package discovery

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

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

func deleteFile(path string) {
	// delete file
	var err = os.Remove(path)
	if isError(err) {
		return
	}

	log.Println("File Deleted")
}

func removeDir(path string) (err error) {
	contents, err := filepath.Glob(path)
	if err != nil {
		return
	}
	for _, item := range contents {
		err = os.RemoveAll(item)
		if err != nil {
			return
		}
	}
	return
}

func createDir(dirName string) error {
	err := os.Mkdir(dirName, 0777)
	if err != nil {
		return err
	}
	return nil
}

// Copy ..
func copy(src, dst string) error {
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
