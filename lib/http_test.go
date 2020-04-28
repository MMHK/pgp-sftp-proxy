package lib

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func getHttpServer() (*HTTPService, error) {
	conf, err := loadConfig()
	if err != nil {
		return nil, err
	}
	
	return NewHTTP(conf), nil
}

func getMultipart(parts map[string]string) (io.Reader, string, error) {
	testUploadFile := getLocalPath("../test/M_Article_Zurich_ca.gif")
	requestReader := new(bytes.Buffer)
	bodyWriter := multipart.NewWriter(requestReader)
	part, err := bodyWriter.CreateFormFile("upload", filepath.Base(testUploadFile))
	if err != nil {
		return nil, "", err
	}
	defer bodyWriter.Close()
	
	testUpload, err := os.Open(testUploadFile)
	if err != nil {
		return nil, "", err
	}
	defer testUpload.Close()
	
	_, err = io.Copy(part, testUpload)
	if err != nil {
		return nil, "", err
	}
	
	for _name, val := range parts {
		w, err := bodyWriter.CreateFormField(_name)
		if err != nil {
			return nil, "", err
		}
		_, err = w.Write([]byte(val))
		if err != nil {
			return nil, "", err
		}
	}
	
	return requestReader, bodyWriter.FormDataContentType(), nil
}

func Test_Encrypt(t *testing.T) {
	httpServer, err := getHttpServer()
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	
	publicKey, err := ioutil.ReadFile(getLocalPath("../test/test-key.pem"))
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	
	requestReader, mime, err := getMultipart(map[string]string{
		"key": string(publicKey),
	})
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	
	req := httptest.NewRequest(http.MethodPost, "/encrypt", requestReader)
	req.Header.Add("Content-Type", mime)
	writer := httptest.NewRecorder()
	
	httpServer.Encrypt(writer, req)
	
	resp := writer.Result()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Response code is %v", resp.StatusCode)
		t.Fail()
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	
	t.Log(string(body))
}

func Test_Upload(t *testing.T) {
	httpServer, err := getHttpServer()
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	
	publicKey, err := ioutil.ReadFile(getLocalPath("../test/test-key.pem"))
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	
	requestReader, mime, err := getMultipart(map[string]string{
		"key": string(publicKey),
	})
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	
	req := httptest.NewRequest(http.MethodPost, "/upload", requestReader)
	req.Header.Add("Content-Type", mime)
	writer := httptest.NewRecorder()
	
	httpServer.Upload(writer, req)
	
	resp := writer.Result()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Response code is %v", resp.StatusCode)
		t.Fail()
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	
	t.Log(string(body))
}
