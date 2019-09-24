package go_fileserver

import "os"

type FileInfo struct {
	Dir      string
	FileName string
}

func NewFile(dir, filename string) *FileInfo {
	return &FileInfo{
		Dir:      dir,
		FileName: filename,
	}
}

func (fi *FileInfo) OpenFile(f *os.File, err error) {
	f, err = os.Open(fi.Dir + fi.FileName)
	defer f.Close()
	if err != nil {
		return
	}
}
