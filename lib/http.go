package lib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

type HTTPService struct {
	config *Config
}
// swagger:response ServiceResult
type ServiceResult struct {
	Status bool        `json:"status"`
	Data   interface{} `json:"data"`
	Error  string      `json:"error"`
}

func NewHTTP(conf *Config) *HTTPService {
	return &HTTPService{
		config: conf,
	}
}

func (this *HTTPService) getHTTPHandler() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/", this.RedirectSample)
	r.HandleFunc("/encrypt", this.Encrypt)
	r.HandleFunc("/upload", this.Upload)
	r.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/",
		http.FileServer(http.Dir(fmt.Sprintf("%s/swagger", this.config.WebRoot)))))
	r.NotFoundHandler = http.HandlerFunc(this.NotFoundHandle)
	
	return r
}

func (this *HTTPService) Start() error {
	log.Info("http service starting")
	log.Infof("Please open http://%s\n", this.config.Listen)
	return http.ListenAndServe(this.config.Listen, this.getHTTPHandler())
}

func (this *HTTPService) NotFoundHandle(writer http.ResponseWriter, request *http.Request) {
	http.Error(writer, "handle not found!", 404)
	this.ResponseError(errors.New("handle not found!"), writer, 404)
}

func (this *HTTPService) RedirectSample(writer http.ResponseWriter, request *http.Request) {
	http.Redirect(writer, request, "/swagger/index.html", 301)
}

func GetMimeType(src *multipart.FileHeader) (string, string, error) {
	e, stat := src.Header["Content-Type"]
	if stat && len(e) > 0 {
		return e[0], src.Filename, nil
	}

	return "", "", errors.New("Not Found MimeInfo")
}

func (this *HTTPService) Upload(writer http.ResponseWriter, request *http.Request) {
	request.ParseMultipartForm(32 << 20)
	uploadFile, header, err := request.FormFile("upload")
	if err != nil {
		log.Error(err)
		this.ResponseError(err, writer, 500)
		return
	}
	defer uploadFile.Close()

	key := request.FormValue("key")

	_, filename, err := GetMimeType(header)
	log.Info("filename:", filename)

	keyReader := strings.NewReader(key)
	remoteFile := filepath.Join(this.config.SFTP.UploadDir, fmt.Sprintf("%s.pgp", filename))
	reader, err := PGP_Encrypt_Reader(uploadFile, keyReader)
	if err != nil {
		log.Error(err)
		this.ResponseError(err, writer, 500)
		return
	}

	ssh := NewStorage(&this.config.SSH)
	err = ssh.PutStream(remoteFile, reader)
	if err != nil {
		log.Error(err)
		this.ResponseError(err, writer, 500)
		return
	}

	this.ResponseJSON(true, writer)
}

func (this *HTTPService) Encrypt(writer http.ResponseWriter, request *http.Request) {
	request.ParseMultipartForm(32 << 20)
	uploadFile, _, err := request.FormFile("upload")
	if err != nil {
		log.Error(err)
		http.Error(writer, err.Error(), 500)
		return
	}
	defer uploadFile.Close()

	key := request.FormValue("key")
	keyReader := strings.NewReader(key)
	reader, err := PGP_Encrypt_Reader(uploadFile, keyReader)
	if err != nil {
		log.Error(err)
		this.ResponseError(err, writer, 500)
		return
	}
	_, err = io.Copy(writer, reader)
	if err != nil {
		log.Error(err)
		return
	}
}

func (this *HTTPService) ResponseError(err error, writer http.ResponseWriter, StatusCode int) {
	server_error := &ServiceResult{Error: err.Error(), Status: false}
	json_str, _ := json.Marshal(server_error)
	writer.Header().Add("Content-Type", "application/json")

	http.Error(writer, string(json_str), StatusCode)
}

func (this *HTTPService) ResponseJSON(src interface{}, writer http.ResponseWriter) {
	serverResult := &ServiceResult{Data: src, Status: true}
	bin, _ := json.Marshal(serverResult)
	reader := bytes.NewReader(bin)

	writer.Header().Add("Content-Type", "application/json")

	io.Copy(writer, reader)
}