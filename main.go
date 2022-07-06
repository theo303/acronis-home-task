package main

import (
	"client/filereader"
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
	fileReaderMgr.Initialize(uri)

	filesToStore, errProcess := fileReaderMgr.Process('A')
	if errProcess != nil {
		fmt.Printf("error during process: %v\n", errProcess)
		return
	}

	for filename, content := range filesToStore {
		os.WriteFile("downloads/"+filename, content, 0644)
	}
	fmt.Printf("exec time: %dms\n", time.Since(startTime).Milliseconds())
}
