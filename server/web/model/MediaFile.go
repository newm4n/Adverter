package model

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	DefaultChunkSize = 100000
)

type TheFile struct {
	ParentPath string
	Name       string
	FilePath   string
	content    []byte
	chunkSize  int
	lastUpdate time.Time
}

func NewTheFile(filePath string) (*TheFile, error) {
	f, err := os.OpenFile(filePath, os.O_RDONLY, os.ModeType)
	if err != nil {
		return nil, err
	}
	inf, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	if inf.IsDir() {
		return nil, fmt.Errorf("%s is not a file", filePath)
	}
	data := make([]byte, inf.Size())
	totRead, err := f.Read(data)
	if totRead != int(inf.Size()) {
		return nil, fmt.Errorf("read size is not equal to actual size read %d != actual %d", totRead, inf.Size())
	}

	lIdx := strings.LastIndex(filePath, string(os.PathSeparator))

	return &TheFile{
		ParentPath: filePath[:lIdx],
		Name:       filePath[lIdx+1:],
		FilePath:   filePath,
		content:    data,
		chunkSize:  DefaultChunkSize,
	}, nil
}

func (tFile *TheFile) GetContent() []byte {
	return tFile.content
}

func (tFile *TheFile) GetChunkCount() int {
	if len(tFile.GetContent())%tFile.chunkSize == 0 {
		return len(tFile.GetContent()) / tFile.chunkSize
	}
	return (len(tFile.GetContent()) / tFile.chunkSize) + 1
}

func (tFile *TheFile) GetHash() (contentHash string, err error) {
	if len(tFile.GetContent()) <= 0 {
		return "", fmt.Errorf("can not hash empty file")
	}
	h := md5.New()
	h.Write(tFile.GetContent())
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (tFile *TheFile) GetBytes(byteFrom, byteTo int) (fromToBytes []byte, fromToHash string, err error) {
	fromToBytes = tFile.GetContent()[byteFrom:byteTo]
	if len(fromToBytes) <= 0 {
		return nil, "", fmt.Errorf("can not hash empty slice")
	}

	h := md5.New()
	h.Write(fromToBytes)
	fromToHash = hex.EncodeToString(h.Sum(nil))
	return fromToBytes, fromToHash, nil
}

func (tFile *TheFile) GetByteOfChunk(chunk int) (chunkBytes []byte, chunkHash string, err error) {
	cStart := tFile.chunkSize * chunk
	cEnd := cStart + tFile.chunkSize
	if cEnd >= len(tFile.GetContent()) {
		cEnd = len(tFile.GetContent())
	}
	chunkBytes = tFile.GetContent()[cStart:cEnd]
	h := md5.New()
	h.Write(chunkBytes)
	chunkHash = hex.EncodeToString(h.Sum(nil))
	return chunkBytes, chunkHash, nil
}

func (tFile *TheFile) GetChunkSize() (currentChunkSize int) {
	return tFile.chunkSize
}

func (tFile *TheFile) SetChunkInfo(newChunkSize int) {
	tFile.chunkSize = newChunkSize
}

type TheDirectory struct {
	ParentPath  string
	Name        string
	DirPath     string
	directories []*TheDirectory
	files       []*TheFile
	lastUpdate  time.Time
}

func NewTheDirectory(path string) (*TheDirectory, error) {
	_, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	lIdx := strings.LastIndex(path, string(os.PathSeparator))
	if lIdx > 0 {
		return &TheDirectory{
			ParentPath:  path[:lIdx],
			Name:        path[lIdx+1:],
			DirPath:     path,
			directories: nil,
			files:       nil,
		}, nil
	}
	return &TheDirectory{
		ParentPath:  "/",
		Name:        path,
		DirPath:     path,
		directories: nil,
		files:       nil,
	}, nil
}

func (tDir *TheDirectory) ListAll() (allFiles []*TheFile, allDir []*TheDirectory, err error) {
	allDir, err = tDir.ListDirectories()
	if err != nil {
		return nil, nil, err
	}
	allFiles, err = tDir.ListFiles()
	if err != nil {
		return nil, nil, err
	}
	return allFiles, allDir, nil
}

func (tDir *TheDirectory) ListFiles() (allFiles []*TheFile, err error) {
	if tDir.lastUpdate.Sub(time.Now()) > (5*time.Minute) || tDir.files == nil {
		fmt.Println("loading files on ", tDir.DirPath)
		allFiles = make([]*TheFile, 0)
		entries, err := os.ReadDir(tDir.DirPath)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if !e.IsDir() {
				fToOpen := fmt.Sprintf("%s%s%s", tDir.DirPath, string(os.PathSeparator), e.Name())
				tf, err := NewTheFile(fToOpen)
				if err != nil {
					fmt.Println("got error for listing file ", e.Name(), ". got ", err.Error())
				} else {
					allFiles = append(allFiles, tf)
				}
			}
		}
		tDir.files = allFiles
		tDir.lastUpdate = time.Now()
		return allFiles, nil
	} else {
		return tDir.files, nil
	}
}

func (tDir *TheDirectory) ListDirectories() (allDir []*TheDirectory, err error) {
	if tDir.lastUpdate.Sub(time.Now()) > (5*time.Minute) || tDir.files == nil {
		fmt.Println("loading dirs on ", tDir.DirPath)
		allDir = make([]*TheDirectory, 0)
		entries, err := os.ReadDir(tDir.DirPath)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() {
				fToOpen := fmt.Sprintf("%s%s%s", tDir.DirPath, string(os.PathSeparator), e.Name())
				tf, err := NewTheDirectory(fToOpen)
				if err != nil {
					fmt.Println("got error for listing dir in ", e.Name(), ". got ", err.Error())
				} else {
					allDir = append(allDir, tf)
				}
			}
		}
		tDir.directories = allDir
		tDir.lastUpdate = time.Now()
		return allDir, nil
	} else {
		return tDir.directories, nil
	}
}

func NewPathInfoFromBase64(pathArg string) (*PathInfo, error) {
	if !strings.Contains(pathArg, ".") {
		return nil, fmt.Errorf("unable to find separator")
	}
	splited := strings.Split(pathArg, ".")
	if len(splited) != 2 {
		return nil, fmt.Errorf("unable to find separator")
	}

	dataBase64Hash := MD5OfBytes([]byte(splited[0]))

	if dataBase64Hash != splited[1] {
		return nil, fmt.Errorf("unable to verify data")
	}
	bPath, err := base64.StdEncoding.DecodeString(splited[0])
	if err != nil {
		return nil, err
	}
	pathInfo := &PathInfo{}
	err = json.Unmarshal(bPath, &pathInfo)
	if err != nil {
		return nil, err
	}
	return pathInfo, nil
}

type PathInfo struct {
	Path string
}

func (pi *PathInfo) ToPathInfoString() string {
	byteData, err := json.Marshal(pi)
	if err != nil {
		panic(fmt.Sprintf("panic. can not marshal path object to json. got %s", err.Error()))
	}
	dataBase64 := base64.StdEncoding.EncodeToString(byteData)
	dataBase64Hash := MD5OfBytes([]byte(dataBase64))
	return fmt.Sprintf("%s.%s", dataBase64, dataBase64Hash)
}

func MD5OfBytes(args []byte) string {
	h := md5.New()
	h.Write(args)
	return hex.EncodeToString(h.Sum(nil))
}
