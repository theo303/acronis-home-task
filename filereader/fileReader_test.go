package filereader

import (
	"bufio"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_fileReader_readByteToChan(t *testing.T) {
	tests := map[string]struct {
		reader    fileReader
		wantValue byte
		wantOpen  bool
	}{
		"read OK": {
			reader: fileReader{
				Reader:   bufio.NewReader(strings.NewReader("a")),
				ReadChan: make(chan byte),
			},
			wantValue: 'a',
			wantOpen:  true,
		},
		"no byte available": {reader: fileReader{
			Reader:   bufio.NewReader(strings.NewReader("")),
			ReadChan: make(chan byte),
		},
			wantValue: 0,
			wantOpen:  false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assertions := assert.New(t)
			go tt.reader.readByteToChan()
			gotValue, gotOk := <-tt.reader.ReadChan
			assertions.Equal(tt.wantValue, gotValue, "byte read")
			assertions.Equal(tt.wantOpen, gotOk, "chan status")
			if gotOk {
				close(tt.reader.ReadChan)
			}
		})
	}
}

func Test_FileReadersManager_Initialize(t *testing.T) {
	tests := map[string]struct {
		fileServerMock func(http.ResponseWriter, *http.Request)
		wantLenReader  int
	}{
		"3 files": {
			fileServerMock: http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/" {
						w.Write([]byte(
							`<pre>
							<p>test</p>
							<a href="0">0</a>
							<a href="1">1</a>
							<a href="2">2</a>
						</pre>`,
						))
					} else {
						w.Write([]byte(`aaa`))
					}
				},
			),
			wantLenReader: 3,
		},
		"no file": {
			fileServerMock: http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/" {
						w.Write([]byte(
							`blablabla`,
						))
					} else {
						w.Write([]byte(`aaa`))
					}
				},
			),
			wantLenReader: 0,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.fileServerMock))
			defer server.Close()

			frm := FileReadersManager{}
			frm.Initialize(server.URL)

			assert.Equal(t, tt.wantLenReader, len(frm.Readers), "number of readers")
		})
	}
}

func Test_FileReadersManager_Process(t *testing.T) {
	content1 := "-a--"
	content2 := "-=b-"
	content3 := "=a--"
	content4 := ""

	tests := map[string]struct {
		char byte
		want map[string][]byte
	}{
		"file 2": {
			char: 'b',
			want: map[string][]byte{
				"2": []byte(content2),
			},
		},
		"file 1 & 3": {
			char: 'a',
			want: map[string][]byte{
				"1": []byte(content1),
				"3": []byte(content3),
			},
		},
		"no file": {
			char: 'c',
			want: map[string][]byte{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			file1Body := io.NopCloser(strings.NewReader(content1))
			file2Body := io.NopCloser(strings.NewReader(content2))
			file3Body := io.NopCloser(strings.NewReader(content3))
			file4Body := io.NopCloser(strings.NewReader(content4))
			frmTest := FileReadersManager{
				Readers: map[string]*fileReader{
					"1": {
						Reader:   bufio.NewReader(file1Body),
						body:     file1Body,
						ReadChan: make(chan byte),
					},
					"2": {
						Reader:   bufio.NewReader(file2Body),
						body:     file2Body,
						ReadChan: make(chan byte),
					},
					"3": {
						Reader:   bufio.NewReader(file3Body),
						body:     file3Body,
						ReadChan: make(chan byte),
					},
					"4": {
						Reader:   bufio.NewReader(file4Body),
						body:     file4Body,
						ReadChan: make(chan byte),
					},
				},
				content: make(map[string][]byte),
			}

			frmTest.Process(tt.char)

			assert.Equal(t, tt.want, frmTest.content, "content map")
			assert.Equal(t, 0, len(frmTest.Readers), "readers map empty")
		})
	}
}
