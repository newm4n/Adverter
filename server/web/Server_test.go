package web

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/newm4n/Adverter/server/web/model"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServerEndpoint(t *testing.T) {
	Router = mux.NewRouter()

	Router.HandleFunc("/path/{b64path}/files", ListFiles).Methods(http.MethodGet)
	Router.HandleFunc("/path/{b64path}/directories", ListDirectories).Methods(http.MethodGet)
	Router.HandleFunc("/path/{b64path}/chunk/info", GetChunkInfo).Methods(http.MethodGet)
	Router.HandleFunc("/path/{b64path}/chunk/{chunkno}", GetChunkData).Methods(http.MethodGet)

	t.Run("Testing file listing", func(t *testing.T) {
		dirToList := "C:\\Users\\ferdi\\Workspace\\Golang\\src\\github.com\\newm4n\\Adverter\\sample"
		pi := model.PathInfo{Path: dirToList}
		str := pi.ToPathInfoString()
		pathToTest := fmt.Sprintf("/path/%s/files", str)

		request, _ := http.NewRequest(http.MethodGet, pathToTest, nil)
		response := httptest.NewRecorder()

		Router.ServeHTTP(response, request)
		if status := response.Code; status != http.StatusOK {
			t.Errorf("Wrong status")
			t.FailNow()
		}
		body, err := io.ReadAll(response.Body)
		assert.NoError(t, err)
		t.Logf("GOT : %s", body)
	})
	t.Run("Testing directory listing", func(t *testing.T) {
		dirToList := "C:\\Users\\ferdi\\Workspace\\Golang\\src\\github.com\\newm4n\\Adverter"
		pi := model.PathInfo{Path: dirToList}
		str := pi.ToPathInfoString()
		pathToTest := fmt.Sprintf("/path/%s/directories", str)

		request, _ := http.NewRequest(http.MethodGet, pathToTest, nil)
		response := httptest.NewRecorder()

		Router.ServeHTTP(response, request)
		if status := response.Code; status != http.StatusOK {
			t.Errorf("Wrong status")
			t.FailNow()
		}
		body, err := io.ReadAll(response.Body)
		assert.NoError(t, err)
		t.Logf("GOT : %s", body)
	})
	t.Run("Testing file info listing", func(t *testing.T) {
		dirToList := "C:\\Users\\ferdi\\Workspace\\Golang\\src\\github.com\\newm4n\\Adverter\\sample\\file_example_MP4_640_3MG.mp4"
		pi := model.PathInfo{Path: dirToList}
		str := pi.ToPathInfoString()
		pathToTest := fmt.Sprintf("/path/%s/chunk/info", str)

		request, _ := http.NewRequest(http.MethodGet, pathToTest, nil)
		response := httptest.NewRecorder()

		Router.ServeHTTP(response, request)
		if status := response.Code; status != http.StatusOK {
			t.Errorf("Wrong status")
			t.FailNow()
		}
		body, err := io.ReadAll(response.Body)
		assert.NoError(t, err)
		t.Logf("GOT : %s", body)
	})
	t.Run("Testing file chunk listing", func(t *testing.T) {
		dirToList := "C:\\Users\\ferdi\\Workspace\\Golang\\src\\github.com\\newm4n\\Adverter\\sample\\file_example_MP4_640_3MG.mp4"
		pi := model.PathInfo{Path: dirToList}
		str := pi.ToPathInfoString()
		for i := 0; i < 32; i++ {
			pathToTest := fmt.Sprintf("/path/%s/chunk/%d", str, i)

			request, _ := http.NewRequest(http.MethodGet, pathToTest, nil)
			response := httptest.NewRecorder()

			Router.ServeHTTP(response, request)
			if status := response.Code; status != http.StatusOK {
				t.Errorf("Wrong status")
				t.FailNow()
			}
			body, err := io.ReadAll(response.Body)
			assert.NoError(t, err)
			res := &ChunkInfoRespond{}
			err = json.Unmarshal(body, &res)
			if err != nil {
				t.Error(err)
				t.FailNow()
			}

			t.Logf("    GOT : %s .. %s (%d bytes) hash : %s", res.Base64[:10], res.Base64[len(res.Base64)-10:], len(res.Base64), res.Hash)
		}
	})
}
