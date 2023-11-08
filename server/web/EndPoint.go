package web

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/newm4n/Adverter/server/web/model"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type DirItemRespond struct {
	Name string
	Path string
	URL  string
}

type ChunkInfoRespond struct {
	Base64 string
	Hash   string
}

type FileInfoRespond struct {
	Name       string
	Path       string
	ParentPath string
	ChunkCount int
	FileHash   string
}

// Router.Handle("/path/{b64path}/files", ListFiles)
func ListFiles(w http.ResponseWriter, r *http.Request) {
	varMap := mux.Vars(r)
	b64Path := varMap["b64path"]
	pathInfo, err := model.NewPathInfoFromBase64(b64Path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("invalid param. got %s", err.Error())))
		return
	}
	tDir, err := model.NewTheDirectory(pathInfo.Path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("invalid param. got %s", err.Error())))
		return
	}
	files, err := tDir.ListFiles()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("invalid param. got %s", err.Error())))
		return
	}
	ret := make([]*DirItemRespond, 0)
	for _, fils := range files {
		pi := &model.PathInfo{
			Path: fils.FilePath,
		}
		d := &DirItemRespond{
			Name: fils.Name,
			Path: fils.FilePath,
			URL:  fmt.Sprintf("/path/%s/files", pi.ToPathInfoString()),
		}
		ret = append(ret, d)
	}
	retBytes, err := json.Marshal(ret)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("invalid param. got %s", err.Error())))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(retBytes)
}

// Router.Handle("/path/{b64path}/directories", ListDirectories)
func ListDirectories(w http.ResponseWriter, r *http.Request) {
	varMap := mux.Vars(r)
	b64Path := varMap["b64path"]
	pathInfo, err := model.NewPathInfoFromBase64(b64Path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("invalid param. got %s", err.Error())))
		return
	}
	tDir, err := model.NewTheDirectory(pathInfo.Path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("invalid param. got %s", err.Error())))
		return
	}
	dirs, err := tDir.ListDirectories()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("invalid param. got %s", err.Error())))
		return
	}
	ret := make([]*DirItemRespond, 0)
	for _, dir := range dirs {
		pi := &model.PathInfo{
			Path: dir.DirPath,
		}
		d := &DirItemRespond{
			Name: dir.Name,
			Path: dir.DirPath,
			URL:  fmt.Sprintf("/path/%s/directories", pi.ToPathInfoString()),
		}
		ret = append(ret, d)
	}
	retBytes, err := json.Marshal(ret)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("invalid param. got %s", err.Error())))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(retBytes)
}

// Router.Handle("/path/{b64path}/chunk/info", GetChunkInfo)
func GetChunkInfo(w http.ResponseWriter, r *http.Request) {
	varMap := mux.Vars(r)
	b64Path := varMap["b64path"]
	pathInfo, err := model.NewPathInfoFromBase64(b64Path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("invalid param. got %s", err.Error())))
		return
	}
	fileFile, err := model.NewTheFile(pathInfo.Path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("invalid param. got %s", err.Error())))
		return
	}

	lIdx := strings.LastIndex(pathInfo.Path, string(os.PathSeparator))

	tFile, err := model.NewTheFile(fileFile.FilePath)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("invalid param. got %s", err.Error())))
		return
	}
	hash, err := tFile.GetHash()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("invalid param. got %s", err.Error())))
		return
	}

	infoResponse := &FileInfoRespond{
		Name:       pathInfo.Path[lIdx+1:],
		ParentPath: pathInfo.Path[:lIdx],
		Path:       pathInfo.Path,
		ChunkCount: tFile.GetChunkCount(),
		FileHash:   hash,
	}

	retBytes, err := json.Marshal(infoResponse)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("invalid param. got %s", err.Error())))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(retBytes)
}

// Router.Handle("/path/{b64path}/chunk/{chunkno}", GetChunkData)
func GetChunkData(w http.ResponseWriter, r *http.Request) {
	varMap := mux.Vars(r)
	b64Path := varMap["b64path"]
	chunkNoStr := varMap["chunkno"]

	pathInfo, err := model.NewPathInfoFromBase64(b64Path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("invalid param. got %s", err.Error())))
		return
	}

	chunkNo, err := strconv.Atoi(chunkNoStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("invalid param. got %s", err.Error())))
		return
	}

	tFile, err := model.NewTheFile(pathInfo.Path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("invalid param. got %s", err.Error())))
		return
	}

	byts, hash, err := tFile.GetByteOfChunk(chunkNo)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("invalid param. got %s", err.Error())))
		return
	}

	cresp := &ChunkInfoRespond{
		Base64: base64.StdEncoding.EncodeToString(byts),
		Hash:   hash,
	}

	retBytes, err := json.Marshal(cresp)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("invalid param. got %s", err.Error())))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(retBytes)
}
