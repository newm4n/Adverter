package logic

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"time"
)

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
	return &TheFile{
		FilePath:  filePath,
		Content:   data,
		ChunkSize: 100000,
	}, nil
}

type TheFile struct {
	FilePath   string
	Content    []byte
	ChunkSize  int
	LastUpdate time.Time
}

func (tFile *TheFile) GetChunkCount() int {
	if len(tFile.Content)%tFile.ChunkSize == 0 {
		return len(tFile.Content) / tFile.ChunkSize
	}
	return (len(tFile.Content) / tFile.ChunkSize) + 1
}

func (tFile *TheFile) GetHash() (contentHash string, err error) {
	if len(tFile.Content) <= 0 {
		return "", fmt.Errorf("can not hash empty file")
	}
	h := md5.New()
	h.Write(tFile.Content)
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (tFile *TheFile) GetBytes(byteFrom, byteTo int) (fromToBytes []byte, fromToHash string, err error) {
	fromToBytes = tFile.Content[byteFrom:byteTo]
	if len(fromToBytes) <= 0 {
		return nil, "", fmt.Errorf("can not hash empty slice")
	}

	h := md5.New()
	h.Write(fromToBytes)
	fromToHash = hex.EncodeToString(h.Sum(nil))
	return fromToBytes, fromToHash, nil
}

func (tFile *TheFile) GetByteOfChunk(chunk int) (chunkBytes []byte, chunkHash string, err error) {
	cStart := tFile.ChunkSize * chunk
	cEnd := cStart + tFile.ChunkSize
	if cEnd >= len(tFile.Content) {
		cEnd = len(tFile.Content)
	}
	chunkBytes = tFile.Content[cStart:cEnd]
	h := md5.New()
	h.Write(chunkBytes)
	chunkHash = hex.EncodeToString(h.Sum(nil))
	return chunkBytes, chunkHash, nil
}

func (tFile *TheFile) GetChunkSize() (currentChunkSize int) {
	return tFile.ChunkSize
}

func (tFile *TheFile) SetChunkInfo(newChunkSize int) {
	tFile.ChunkSize = newChunkSize
}

func NewTheDirectory(path string) (*TheDirectory, error) {
	_, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	return &TheDirectory{
		DirPath:     path,
		directories: nil,
		files:       nil,
	}, nil
}

type TheDirectory struct {
	DirPath     string
	directories []*TheDirectory
	files       []*TheFile
	lastUpdate  time.Time
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
