package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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
	r.HandleFunc("/", this.RedirectSample)
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
	file, header, err := request.FormFile("upload")
	if err != nil {
		log.Error(err)
		this.ResponseError(err, writer, 500)
		return
	}
	defer file.Close()

	key := request.FormValue("key")
	deploy_type := request.FormValue("deploy")

	var src_data []byte
	mimeType, filename, err := GetMimeType(header)
	log.Info("filename:", filename)
	if err == nil && strings.Contains(mimeType, "image") {
		//convert to pdf
		reader, err := getPDFBytes(file, this.config.TempPath)
		if err != nil {
			log.Error(err)
			this.ResponseError(err, writer, 500)
			return
		}
		src_data, err = ioutil.ReadAll(reader)
		if err != nil {
			log.Error(err)
			this.ResponseError(err, writer, 500)
			return
		}
		filename = filename + ".pdf"
	} else {
		src_data, err = ioutil.ReadAll(file)
		if err != nil {
			log.Error(err)
			this.ResponseError(err, writer, 500)
			return
		}
	}

	key_reader := strings.NewReader(key)
	remoteFile := path.Join(this.config.GetDeployPath(deploy_type), filename+".pgp")
	bin, err := PGP_Encrypt(src_data, key_reader)
	if err != nil {
		log.Error(err)
		this.ResponseError(err, writer, 500)
		return
	}

	ssh := NewSSHClient(&this.config.SSH)
	err = ssh.Put(remoteFile, strings.NewReader(bin))
	if err != nil {
		log.Error(err)
		this.ResponseError(err, writer, 500)
		return
	}

	fmt.Fprintf(writer, "{\"status\":1}")
}

func (this *HTTPService) Encrypt(writer http.ResponseWriter, request *http.Request) {
	request.ParseMultipartForm(32 << 20)
	file, header, err := request.FormFile("upload")
	if err != nil {
		log.Error(err)
		http.Error(writer, err.Error(), 500)
		return
	}
	defer file.Close()

	key := request.FormValue("key")
	var src_data []byte
	mimeType, _, err := GetMimeType(header)
	if err == nil && strings.Contains(mimeType, "image") {
		//convert to pdf
		reader, err := getPDFBytes(file, this.config.TempPath)
		if err != nil {
			log.Error(err)
			this.ResponseError(err, writer, 500)
			return
		}

		src_data, err = ioutil.ReadAll(reader)
		if err != nil {
			log.Error(err)
			this.ResponseError(err, writer, 500)
			return
		}
	} else {
		src_data, err = ioutil.ReadAll(file)
		if err != nil {
			log.Error(err)
			this.ResponseError(err, writer, 500)
			return
		}
	}
	keyReader := strings.NewReader(key)
	encodeEntry, err := PGP_Encrypt(src_data, keyReader)
	if err != nil {
		log.Error(err)
		this.ResponseError(err, writer, 500)
		return
	}
	fmt.Fprintf(writer, encodeEntry)
}

func (this *HTTPService) ResponseError(err error, writer http.ResponseWriter, StatusCode int) {
	server_error := &ServiceResult{Error: err.Error(), Status: false}
	json_str, _ := json.Marshal(server_error)
	writer.Header().Add("Content-Type", "application/json")

	http.Error(writer, string(json_str), StatusCode)
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

	fmt.Fprintf(writer, "{\"status\":1}")
	return
}
