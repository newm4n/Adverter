package logic

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"testing"
)

//base := "C:\\Users\\ferdi\\Workspace\\Golang\\src\\github.com\\newm4n\\Adverter\\sample"

func TestListingDirectory(t *testing.T) {

	//tDir, err := NewTheDirectory("C:\\Users\\ferdi\\Workspace\\Golang\\src\\github.com\\newm4n\\Adverter")
	tDir, err := NewTheDirectory("C:\\Users\\ferdi\\Workspace\\Golang\\src\\github.com\\newm4n\\Adverter\\sample")
	assert.NoError(t, err)
	assert.NotNil(t, tDir)

	dirs, err := tDir.ListDirectories()
	assert.NoError(t, err)
	t.Log("Dir size ", tDir.DirPath, " : ", len(dirs))

	files, err := tDir.ListFiles()
	assert.NoError(t, err)
	t.Log("File size ", tDir.DirPath, " : ", len(files))
	for fi, f := range files {
		t.Log("#", fi, " : ", f.FilePath, " ", len(f.Content), " bytes. ", f.GetChunkCount(), " chunks where ", f.ChunkSize, " bytes each chunk. ")
		h, err := f.GetHash()
		if err != nil {
			t.Log("     hash error ", err.Error())
		}
		t.Log("   Hash : " + h)
		bb := bytes.Buffer{}
		for ck := 0; ck < f.GetChunkCount(); ck++ {
			cBytes, cHash, err := f.GetByteOfChunk(ck)
			if err != nil {
				t.Log("       chunk byte #", ck, " error ", err.Error())
			} else {
				bb.Write(cBytes)
				t.Log("       chunk #", ck, " : ", len(cBytes), " bytes. Chunk hash ", cHash)
			}
		}

		h2 := md5.New()
		h2.Write(bb.Bytes())
		h2sum := hex.EncodeToString(h2.Sum(nil))
		t.Log("       Hash Total : " + h2sum)

		if h != h2sum {
			t.FailNow()
		}
	}
}
