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
	err := fileReaderMgr.Initialize(uri)
	if err != nil {
		panic(fmt.Errorf("error initializing fileReaderManager: %v", err))
	}

	fileReaderMgr.Process('A')

	err = fileReaderMgr.StoreFiles("downloads")
	if err != nil {
		panic(fmt.Errorf("error storing files: %v", err))
	}
	fmt.Printf("exec time: %dms\n", time.Since(startTime).Milliseconds())
}
