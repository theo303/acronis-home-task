package filereader

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/exp/slices"
)

type fileReader struct {
	*bufio.Reader
	body     io.ReadCloser
	ReadChan chan byte
}

func new(uri string) (*fileReader, error) {
	resp, getErr := http.Get(uri)
	if getErr != nil {
		return nil, fmt.Errorf("issuing http get : %w", getErr)
	}

	return &fileReader{
		Reader:   bufio.NewReader(resp.Body),
		body:     resp.Body,
		ReadChan: make(chan byte),
	}, nil
}

func (fr *fileReader) readByteToChan() {
	byteRead, readErr := fr.ReadByte()
	if readErr != nil {
		close(fr.ReadChan)
		return
	}
	fr.ReadChan <- byteRead
}

func (fr *fileReader) close() {
	fr.body.Close()
	close(fr.ReadChan)
}

// FileReadersManager is used to manage all file readers
type FileReadersManager struct {
	Readers map[string]*fileReader

	content map[string][]byte
}

// Initialize a file reader for each file available on the uri
func (frm *FileReadersManager) Initialize(uri string) error {
	filesList, listErr := listFiles(uri)
	if listErr != nil {
		return fmt.Errorf("listing files available on server : %w", listErr)
	}

	frm.Readers = make(map[string]*fileReader)
	frm.content = make(map[string][]byte)

	var err error
	errStr := strings.Builder{}
	for _, fileLink := range filesList {
		frm.Readers[fileLink], err = new(fmt.Sprintf("%s/%s", uri, fileLink))
		if err != nil {
			errStr.WriteString(fmt.Sprintf("%s: %v - ", fileLink, err))
			continue
		}
	}
	if errStr.Len() != 0 {
		return fmt.Errorf("creating reader for file: %s", errStr.String())
	}
	return nil
}

// Process returns the content of the file(s) that contained the char on the earliest position
func (frm *FileReadersManager) Process(char byte) {
	charFound := false
	for len(frm.Readers) != 0 {
		fileWithChar := []string{}
		for file, reader := range frm.Readers {
			go reader.readByteToChan()
			byteRead, ok := <-reader.ReadChan
			if !ok {
				// file was read entirely before char was found
				if !charFound {
					delete(frm.content, file)
				}
				delete(frm.Readers, file)
				continue
			}
			if byteRead == char {
				fileWithChar = append(fileWithChar, file)
			}
			frm.content[file] = append(frm.content[file], byteRead)
		}
		if len(fileWithChar) != 0 {
			charFound = true
			for filename := range frm.Readers {
				if !slices.Contains(fileWithChar, filename) {
					frm.remove(filename)
				}
			}
		}
		// char has been found on previous byte read, now we can finish reading files without checks
		if charFound {
			frm.finishReading()
		}
	}
}

func (frm *FileReadersManager) finishReading() {
	for len(frm.Readers) != 0 {
		for file, reader := range frm.Readers {
			go reader.readByteToChan()
			byteRead, ok := <-reader.ReadChan
			if !ok {
				delete(frm.Readers, file)
				continue
			}
			frm.content[file] = append(frm.content[file], byteRead)
		}
	}
}

func (frm *FileReadersManager) remove(file string) {
	frm.Readers[file].close()
	delete(frm.Readers, file)
	delete(frm.content, file)
}

func (frm *FileReadersManager) StoreFiles(path string) error {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		errDir := os.Mkdir(path, os.ModePerm)
		if errDir != nil {
			return fmt.Errorf("creating directory: %w", errDir)
		}
	}

	for filename, content := range frm.content {
		errWrite := os.WriteFile("downloads/"+filename, content, 0644)
		if errWrite != nil {
			return fmt.Errorf("writing file %s: %w", filename, errWrite)
		}
		delete(frm.content, filename)
	}
	return nil
}

func listFiles(uri string) ([]string, error) {
	resp, getErr := http.Get(uri)
	if getErr != nil {
		return nil, fmt.Errorf("issuing http get : %w", getErr)
	}

	document, docErr := goquery.NewDocumentFromReader(resp.Body)
	if docErr != nil {
		return nil, fmt.Errorf("creating goquery document : %w", docErr)
	}

	filesList := []string{}
	document.Find("a").Each(func(_ int, element *goquery.Selection) {
		fileLink, exists := element.Attr("href")
		if exists {
			filesList = append(filesList, fileLink)
		}
	})
	return filesList, nil
}
