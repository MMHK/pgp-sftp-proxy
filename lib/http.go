package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path"
	"strings"
	
	"github.com/gorilla/mux"
)

type HTTPService struct {
	config *Config
}

type ServiceResult struct {
	Status bool   `json:"status"`
	Error  string `json:"error"`
}

// swagger:model
type MultipleBody struct {
	// in:body

	// upload files
	// required: true
	Files []*ZurichFile `json:"files"`
	// PGP public key
	// required: true
	PGPKey string `json:"key"`
	// sftp remote save folder
	// required: true
	// enum: dev, pro, test
	ENV string `json:"env"`
	// notify URL
	NotifyURL string `json:"notify"`
}

func NewHTTP(conf *Config) *HTTPService {
	return &HTTPService{
		config: conf,
	}
}

func (this *HTTPService) getHTTPHandler() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/", this.RedirectSwagger)
	r.HandleFunc("/encrypt", this.Encrypt)
	r.HandleFunc("/upload", this.Upload)
	r.HandleFunc("/multiple/upload", this.Multiple)
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

func (this *HTTPService) RedirectSwagger(writer http.ResponseWriter, request *http.Request) {
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
	err := request.ParseMultipartForm(32 << 20)
	if err != nil {
		log.Error(err)
		this.ResponseError(err, writer, 500)
		return
	}
	
	var reader io.Reader
	
	file, header, err := request.FormFile("upload")
	if err != nil {
		log.Error(err)
		this.ResponseError(err, writer, 500)
		return
	}
	defer file.Close()
	reader = file

	key := request.FormValue("key")
	deploy_type := request.FormValue("deploy")

	mimeType, filename, err := GetMimeType(header)
	log.Info("filename:", filename)
	if err == nil && strings.Contains(mimeType, "image") {
		//convert to pdf
		reader, err = getPDFBytes(file, this.config.TempPath)
		if err != nil {
			log.Error(err)
			this.ResponseError(err, writer, 500)
			return
		}
		filename = filename + ".pdf"
	}

	keyReader := strings.NewReader(key)
	remoteFile := path.Join(this.config.GetDeployPath(deploy_type), filename+".pgp")
	helper, err := NewPGPHelper(keyReader)
	if err != nil {
		log.Error(err)
		this.ResponseError(err, writer, 500)
		return
	}
	buffer, err := helper.Encrypt(reader)
	if err != nil {
		log.Error(err)
		this.ResponseError(err, writer, 500)
		return
	}
	ssh := NewSSHClient(&this.config.SSH)
	err = ssh.Put(remoteFile, buffer)
	if err != nil {
		log.Error(err)
		this.ResponseError(err, writer, 500)
		return
	}

	this.ResponseJSON(ServiceResult{Status:true}, writer, 200)
}

func (this *HTTPService) Encrypt(writer http.ResponseWriter, request *http.Request) {
	err := request.ParseMultipartForm(32 << 20)
	if err != nil {
		log.Error(err)
		this.ResponseError(err, writer, 500)
		return
	}
	
	var reader io.Reader
	file, header, err := request.FormFile("upload")
	if err != nil {
		log.Error(err)
		http.Error(writer, err.Error(), 500)
		return
	}
	defer file.Close()
	reader = file

	key := request.FormValue("key")
	mimeType, _, err := GetMimeType(header)
	if err == nil && strings.Contains(mimeType, "image") {
		//convert to pdf
		reader, err = getPDFBytes(file, this.config.TempPath)
		if err != nil {
			log.Error(err)
			this.ResponseError(err, writer, 500)
			return
		}
	}
	keyReader := strings.NewReader(key)
	helper, err := NewPGPHelper(keyReader)
	if err != nil {
		log.Error(err)
		this.ResponseError(err, writer, 500)
		return
	}
	buffer, err := helper.Encrypt(reader)
	if err != nil {
		log.Error(err)
		this.ResponseError(err, writer, 500)
		return
	}
	
	_, err = io.Copy(writer, buffer)
	if err != nil {
		log.Error(err)
		this.ResponseError(err, writer, 500)
		return
	}
}

func (this *HTTPService) ResponseError(err error, writer http.ResponseWriter, StatusCode int) {
	this.ResponseJSON(err, writer, 500)
}

func (this *HTTPService) ResponseJSON(obj interface{}, writer http.ResponseWriter, StatusCode int) {
	jsonString, _ := json.Marshal(obj)
	writer.Header().Add("Content-Type", "application/json")
	
	fmt.Fprint(writer, string(jsonString))
}

func (this *HTTPService) Multiple(writer http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)
	var reqBody MultipleBody
	err := decoder.Decode(&reqBody)
	if err != nil {
		log.Error(err)
		this.ResponseError(errors.New("decode request body error"), writer, 500)
		return
	}

	if len(reqBody.Files) <= 0 {
		this.ResponseError(errors.New("File list empty"), writer, 500)
		return
	}

	if len(reqBody.PGPKey) <= 0 {
		this.ResponseError(errors.New("PGPKey empty"), writer, 500)
		return
	}

	if len(reqBody.ENV) <= 0 {
		this.ResponseError(errors.New("ENV empty"), writer, 500)
		return
	}

	z := NewZurich(this.config, reqBody.Files, reqBody.PGPKey, reqBody.ENV, reqBody.NotifyURL)
	go z.Process()

	this.ResponseJSON(ServiceResult{Status:true}, writer, 200)
}
