package main

import (
	"client/filereader"
	"errors"
	"fmt"
	"os"
	"time"
)

func main() {
	startTime := time.Now()
	uri := "http://localhost:8080"
	if len(os.Args) > 2 {
		uri = fmt.Sprintf("http://%s:%s", os.Args[1], os.Args[2])
	}

	fileReaderMgr := filereader.FileReadersManager{}
	err := fileReaderMgr.Initialize(uri)
	if err != nil {
		panic(fmt.Errorf("error initializing fileReaderManager: %v", err))
	}

	filesToStore := fileReaderMgr.Process('A')

	err = storeFiles(filesToStore, "downloads")
	if err != nil {
		panic(fmt.Errorf("error storing files: %v", err))
	}
	fmt.Printf("exec time: %dms\n", time.Since(startTime).Milliseconds())
}

func storeFiles(filesToStore map[string][]byte, path string) error {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		errDir := os.Mkdir(path, os.ModePerm)
		if errDir != nil {
			return fmt.Errorf("creating directory: %w", errDir)
		}
	}

	for filename, content := range filesToStore {
		errWrite := os.WriteFile("downloads/"+filename, content, 0644)
		if errWrite != nil {
			return fmt.Errorf("writing file %s: %w", filename, errWrite)
		}
	}
	return nil
}
