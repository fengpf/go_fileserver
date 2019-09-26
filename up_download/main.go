package main

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	//MB represents a megabyte
	MB           = 1 << 20
	serverKey    = "bleashup-file-server"
	photosFolder = "/Users/fpf/Downloads/"
)

type file struct {
	File       multipart.File
	FileHeader *multipart.FileHeader
}

func main() {
	var routes = mux.NewRouter()
	routes.HandleFunc("/photo/get/{filename}", getPhoto).Methods("GET")
	routes.HandleFunc("/photo/save", savePhoto).Methods("POST")
	http.Handle("/", routes)
	log.Fatal(http.ListenAndServe(":8855", nil))

}

func processGet(w http.ResponseWriter, r *http.Request, parentDir string) {
	var vars = mux.Vars(r)
	fileName := parentDir + vars["filename"]
	if fileName == "" {
		http.Error(w, "Get 'file' not specified in url.", 400)
		return
	}
	fmt.Println("client requests: " + fileName)

	openFile, err := os.Open(fileName)
	defer openFile.Close()
	if err != nil {
		fmt.Printf("os.Open error(%v)", err)
		return
	}

	fileStat, err := openFile.Stat()
	if err != nil {
		fmt.Printf("openFile.Stat error(%v)", err)
		return
	}

	fileHeader := make([]byte, 32)
	openFile.Read(fileHeader)

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.Header().Set("Content-Type", http.DetectContentType(fileHeader))
	w.Header().Set("Content-Length", strconv.FormatInt(fileStat.Size(), 10))

	pr, pw := io.Pipe()
	defer func() {
		pw.Close()
		pr.Close()
	}()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		if _, err = io.Copy(pw, openFile); err != nil {
			return
		}
	}()

	go func() {
		defer wg.Done()

		openFile.Seek(0, 0) //从文件头重新读取数据
		if _, err = io.Copy(w, pr); err != nil {
			return
		}
	}()

	wg.Wait()
	return
}

func retrievePhotoInfo(w http.ResponseWriter, r *http.Request) (file, error) {
	fileInstance := file{}
	var err error
	if err = r.ParseMultipartForm(5 * MB); err != nil {
		fmt.Println(err.Error())
		return file{}, err
	}
	fileInstance.File, fileInstance.FileHeader, err = r.FormFile("file")
	fmt.Println(fileInstance.FileHeader.Filename, "********************")
	if err != nil {
		fmt.Println(err.Error())
		return file{}, err
	}
	return fileInstance, nil

}

func generateID() string {
	return uuid.New().String()
}

func savePhoto(w http.ResponseWriter, r *http.Request) {
	processSave(w, r, photosFolder)
}

func getPhoto(w http.ResponseWriter, r *http.Request) {
	processGet(w, r, photosFolder)
}

func processSave(w http.ResponseWriter, r *http.Request, parentDir string) {
	fileInstance := file{}
	fileInstance, err := retrievePhotoInfo(w, r)
	defer fileInstance.File.Close()
	if err != nil {
		fmt.Println(err)
	}

	ID := generateID() + "_" + generateID() + "_" + generateID()
	filenames := strings.Split(fileInstance.FileHeader.Filename, ".")
	fmt.Println(filenames[len(filenames)-1])

	var file *os.File
	images, err := filepath.Glob(parentDir + ID + ".*")
	if err != nil {
		fmt.Println(err.Error())
	}

	if images != nil {
		file, err = os.Create(parentDir + ID + "." + filenames[len(filenames)-1])
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	file, err = os.Create(parentDir + ID + "." + filenames[len(filenames)-1])
	if err != nil {
		fmt.Println(err.Error())
	}

	io.Copy(file, fileInstance.File)

	if err != nil {
		fmt.Println(err.Error())
	}

	w.Write([]byte(ID + "." + filenames[len(filenames)-1]))
}
