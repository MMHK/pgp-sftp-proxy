package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type HTTPService struct {
	config *Config
}

type ServiceResult struct {
	Status bool   `json:"status"`
	Error  string `json:"error"`
}

type MultipleBody struct {
	Files     map[string]string `json:"files"`
	PGPKey    string            `json:"key"`
	ENV       string            `json:"env"`
	NotifyURL string            `json:"notify"`
}

func NewHTTP(conf *Config) *HTTPService {
	return &HTTPService{
		config: conf,
	}
}

func (s *HTTPService) Start() {
	r := mux.NewRouter()
	r.HandleFunc("/", s.RedirectSample)
	r.HandleFunc("/encrypt", s.Encrypt)
	r.HandleFunc("/upload", s.Upload)
	r.HandleFunc("/multiple/upload", s.Multiple)
	r.PathPrefix("/sample/").Handler(http.StripPrefix("/sample/",
		http.FileServer(http.Dir(fmt.Sprintf("%s/sample", s.config.WebRoot)))))
	r.NotFoundHandler = http.HandlerFunc(s.NotFoundHandle)

	InfoLogger.Println("http service starting")
	InfoLogger.Printf("Please open http://%s\n", s.config.Listen)
	http.ListenAndServe(s.config.Listen, r)
}

func (s *HTTPService) NotFoundHandle(writer http.ResponseWriter, request *http.Request) {
	http.Error(writer, "handle not found!", 404)
	s.ResponseError(errors.New("handle not found!"), writer, 404)
}

func (s *HTTPService) RedirectSample(writer http.ResponseWriter, request *http.Request) {
	http.Redirect(writer, request, "/sample/index.html", 301)
}

func GetMimeType(src *multipart.FileHeader) (string, string, error) {
	e, stat := src.Header["Content-Type"]
	if stat && len(e) > 0 {
		return e[0], src.Filename, nil
	}

	return "", "", errors.New("Not Found MimeInfo")
}

func (s *HTTPService) Upload(writer http.ResponseWriter, request *http.Request) {
	request.ParseMultipartForm(32 << 20)
	file, header, err := request.FormFile("upload")
	if err != nil {
		ErrLogger.Println(err)
		s.ResponseError(err, writer, 500)
		return
	}
	defer file.Close()

	key := request.FormValue("key")
	deploy_type := request.FormValue("deploy")

	var src_data []byte
	mimeType, filename, err := GetMimeType(header)
	InfoLogger.Println("filename:", filename)
	if err == nil && strings.Contains(mimeType, "image") {
		//convert to pdf
		src_data, err = SavePDF(file, filename, s.config.TempPath)
		if err != nil {
			ErrLogger.Println(err)
			s.ResponseError(err, writer, 500)
			return
		}
		filename = filename + ".pdf"
	} else {
		src_data, err = ioutil.ReadAll(file)
		if err != nil {
			ErrLogger.Println(err)
			s.ResponseError(err, writer, 500)
			return
		}
	}

	key_reader := strings.NewReader(key)
	filename = path.Join(s.config.TempPath, filename+".pgp")
	err = PGP_Encrypt_File(src_data, key_reader, filename)
	if err != nil {
		ErrLogger.Println(err)
		s.ResponseError(err, writer, 500)
		return
	}

	ssh := NewSSHClient(&s.config.SSH)
	err = ssh.UploadFile(filename, s.config.GetDeployPath(deploy_type))
	if err != nil {
		ErrLogger.Println(err)
		s.ResponseError(err, writer, 500)
		return
	}
	//remove uploaded file
	defer func() {
		time.AfterFunc(time.Second*2, func() {
			err = os.Remove(filename)
			if err != nil {
				ErrLogger.Println(err)
			}
		})
	}()
	fmt.Fprintf(writer, "{\"status\":1}")
}

func (s *HTTPService) Encrypt(writer http.ResponseWriter, request *http.Request) {
	request.ParseMultipartForm(32 << 20)
	file, header, err := request.FormFile("upload")
	if err != nil {
		ErrLogger.Println(err)
		http.Error(writer, err.Error(), 500)
		return
	}
	defer file.Close()

	key := request.FormValue("key")
	var src_data []byte
	mimeType, filename, err := GetMimeType(header)
	if err == nil && strings.Contains(mimeType, "image") {
		//convert to pdf
		src_data, err = SavePDF(file, filename, s.config.TempPath)
		if err != nil {
			ErrLogger.Println(err)
			s.ResponseError(err, writer, 500)
			return
		}
	} else {
		src_data, err = ioutil.ReadAll(file)
		if err != nil {
			ErrLogger.Println(err)
			s.ResponseError(err, writer, 500)
			return
		}
	}

	key_reader := strings.NewReader(key)
	encodeEntry, err := PGP_Encrypt(src_data, key_reader)
	if err != nil {
		ErrLogger.Println(err)
		s.ResponseError(err, writer, 500)
		return
	}
	fmt.Fprintf(writer, encodeEntry)
}

func (s *HTTPService) ResponseError(err error, writer http.ResponseWriter, StatusCode int) {
	server_error := &ServiceResult{Error: err.Error(), Status: false}
	json_str, _ := json.Marshal(server_error)
	http.Error(writer, string(json_str), StatusCode)
}

func (s *HTTPService) Multiple(writer http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)
	var reqBody MultipleBody
	err := decoder.Decode(&reqBody)
	if err != nil {
		ErrLogger.Println(err)
		s.ResponseError(errors.New("decode request body error"), writer, 500)
		return
	}

	if len(reqBody.Files) <= 0 {
		s.ResponseError(errors.New("File list empty"), writer, 500)
		return
	}

	if len(reqBody.PGPKey) <= 0 {
		s.ResponseError(errors.New("PGPKey empty"), writer, 500)
		return
	}

	if len(reqBody.ENV) <= 0 {
		s.ResponseError(errors.New("ENV empty"), writer, 500)
		return
	}

	files := make([]*ZurichFile, len(reqBody.Files))
	index := 0
	for filename, url := range reqBody.Files {
		files[index] = &ZurichFile{
			Name: filename,
			Url:  url,
		}
		index++
	}

	z := NewZurich(s.config, files, reqBody.PGPKey, reqBody.ENV, reqBody.NotifyURL)
	go z.Process()

	fmt.Fprintf(writer, "{\"status\":1}")
	return
}
