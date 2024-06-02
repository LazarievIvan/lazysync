package filesystem

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	"github.com/tidwall/gjson"
	"io"
	"lazysync/application/service"
	"lazysync/application/web"
	"lazysync/modules/filesystem/cmd"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

const ID = "filesystem"

const port = ":3000"

const host = "localhost"

type FileSync struct {
	id            string
	Configuration FileSyncConfig
}

type FileSyncConfig struct {
	Files []string // Files list.
}

type FileSyncObject struct {
	DownloadUrl string
	Files       []string
}

type FileSyncArgs struct{}

type FileSyncRequest struct {
	Method string         `json:"method"`
	Params []FileSyncArgs `json:"params"`
	Id     string         `json:"id"`
}

type FileSyncResponse struct {
	FileName     string `json:"filename"`
	FileContents string `json:"contents"`
}

func Init() *FileSync {
	return &FileSync{id: ID, Configuration: FileSyncConfig{}}
}

func (f *FileSync) GetId() string {
	return f.id
}

func (f *FileSync) SetupModule() {
	f.Configuration.Files = cmd.Setup()
}

func (f *FileSync) GetConfigurationValues() interface{} {
	return f.Configuration
}

func (f *FileSync) SetConfiguration(configuration interface{}) {
	var filelist []string
	for _, value := range configuration.(map[string]interface{}) {
		for _, filename := range value.([]interface{}) {
			filelist = append(filelist, filename.(string))
		}
	}
	f.Configuration.Files = filelist
}

func (f *FileSync) Sync() service.SyncObject {
	actionId := uuid.New().String()
	path := "/download/" + actionId
	fullUrl := web.DefaultUrl + path
	syncResponse := FileSyncObject{
		DownloadUrl: fullUrl,
		Files:       f.Configuration.Files,
	}
	return &syncResponse
}

func (f *FileSync) ExecuteCommands(object service.SyncObject) {
	fileSyncObject := object.(*FileSyncObject)
	var wg sync.WaitGroup
	wg.Add(len(fileSyncObject.Files))
	for _, file := range fileSyncObject.Files {
		go DoDownload(fileSyncObject.DownloadUrl, file, &wg)
	}
	wg.Wait()
}

func (f *FileSync) GetSyncObjectInstance() service.SyncObject {
	return new(FileSyncObject)
}

func (f *FileSyncObject) ParseResponse(jsonResponse string) {
	f.DownloadUrl = gjson.Get(jsonResponse, "DownloadUrl").String()
	files := gjson.Get(jsonResponse, "Files").Array()
	for _, file := range files {
		f.Files = append(f.Files, file.String())
	}
}

func DoDownload(downloadUrl string, filepath string, wg *sync.WaitGroup) {
	defer wg.Done()
	tokens := strings.Split(filepath, "/")
	fileName := tokens[len(tokens)-1]
	fmt.Println("Downloading", fileName, "to", fileName)
	output, err := os.Create(fileName)
	if err != nil {
		log.Fatal("Error while creating", fileName, "-", err)
	}
	defer output.Close()
	arguments := FileSyncArgs{}
	request := newDownloadFilesRequest()
	request.Params = append(request.Params, arguments)
	request.Id = "3"
	jsonData, err := json.Marshal(request)
	if err != nil {
		panic(err)
	}
	fullUrl := downloadUrl + "/" + fileName
	resp, err := web.SendJsonRequest(http.MethodPost, fullUrl, jsonData)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	filecontents := gjson.GetBytes(responseBody, "result.contents")
	decoded, err := base64.StdEncoding.DecodeString(filecontents.String())
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(output, bytes.NewReader(decoded))
	if err != nil {
		panic(err)
	}
	fmt.Println("Downloaded ", fileName)
}

func newDownloadFilesRequest() *FileSyncRequest {
	request := new(FileSyncRequest)
	request.Method = "FileSync.HandleDownload"
	return request
}

func (f *FileSync) RegisterAsWebService(router *mux.Router, server *rpc.Server) {
	path := "/download/{actionId}/{filename}"
	router.Handle(path, server)
}

func (f *FileSync) HandleDownload(r *http.Request, args *FileSyncArgs, reply *FileSyncResponse) error {
	params := mux.Vars(r)
	filename := params["filename"]
	if filename == "" {
		return errors.New("requested file not found")
	}
	var response FileSyncResponse
	filesList := f.Configuration.Files
	for _, filePath := range filesList {
		tokens := strings.Split(filePath, "/")
		enabledFilename := tokens[len(tokens)-1]
		if filename != enabledFilename {
			continue
		}
		fileContents, err := os.ReadFile(filePath)
		encodedbytes := base64.StdEncoding.EncodeToString(fileContents)
		if err != nil {
			return err
		}
		response.FileContents = encodedbytes
	}
	response.FileName = filename
	*reply = response
	return nil
}
