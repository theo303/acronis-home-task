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

	filesToStore := fileReaderMgr.Process('A')

	for filename, content := range filesToStore {
		os.WriteFile("downloads/"+filename, content, 0644)
	}
	fmt.Printf("exec time: %dms\n", time.Since(startTime).Milliseconds())
}
