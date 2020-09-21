package lib

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func getHttpServer() (*HTTPService, error) {
	conf, err := loadConfig()
	if err != nil {
		return nil, err
	}
	
	return NewHTTP(conf), nil
}

func getMultipart(parts map[string]string) (io.Reader, string, error) {
	testUploadFile := getLocalPath("../test/temp/dahsing-SCHEDULE-sample.pdf")
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

func Test_Multiple(t *testing.T) {
	requestBodyFile := getLocalPath("../test/request.json")
	requestReader, err := os.Open(requestBodyFile)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	defer requestReader.Close()
	
	httpServer, err := getHttpServer()
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	
	req := httptest.NewRequest(http.MethodPost, "/multiple/upload", requestReader)
	
	writer := httptest.NewRecorder()
	
	httpServer.Multiple(writer, req)
	
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
	
	time.Sleep(time.Second * 20)
	
	t.Log(string(body))
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
	
	publicKey, err := ioutil.ReadFile(getLocalPath("../test/temp/dahsing-public-key.pem"))
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

func Test_DecodeBody(t *testing.T) {
	source := `{"files":{"M_Article_Zurich_2.gif":"https:\/\/s3-ap-southeast-1.amazonaws.com\/s3.jetso.com\/asset\/images\/M_Article_Zurich_2.gif","M_Article_Zurich_4.jpg":"https:\/\/s3-ap-southeast-1.amazonaws.com\/s3.jetso.com\/asset\/images\/M_Article_Zurich_4.jpg","M_Article_Zurich_ca.gif":"https:\/\/s3-ap-southeast-1.amazonaws.com\/s3.jetso.com\/asset\/images\/M_Article_Zurich_ca.gif"},"key":"-----BEGIN PGP PUBLIC KEY BLOCK-----\r\nVersion: GnuPG v2\r\n\r\nmQENBFV2aQkBCADuEi0WB\/VeHp2zo\/6XRnX6uLbIyKQszo0gW6Ek4WGdTvovX\/9r\r\nh6qNx++pcLmT8wmuwvMMIyvsNEt5eKsWSgjJZfSqwo2uMYePpz2ZjruC+eGzONS5\r\nnWBbmScnmGphlLXnW8OpOb2JFqiZRj8Rv+UEUy39DsFiwsNBRkYzWgbX6yI7YgNH\r\nRZxcCWvhZZrDbBSDlhzzSFQttVS+PchvI1rXkgbO5igopsolj86LnB0HnZqlivNE\r\naQ1xxfTKPv9tKm3DeZqEPdbpBkxBdrqDEye9Gjq06wgJQ68bIzwqAAFuCKKWfeCg\r\nCclw3MaVTXX5wuwl4V8mqVvkMOUt9Qkli149ABEBAAG0J0RSSVZFUl9VQVRfS0VZ\r\nIDx0ZXJyYW5jZUBkcml2ZXIuY29tLmhrPokBOQQTAQgAIwUCVXZpCQIbDwcLCQgH\r\nAwIBBhUIAgkKCwQWAgMBAh4BAheAAAoJENxsjOiA47hHuxIH\/3Y8DgLiM0oD6opP\r\nN1Wwnd5f9\/J3is9WlaKuxGP6iDjHKfTf2Bcwm5AC1+XosW6HSrd7g9JiubG6Cvsz\r\nkI\/voFVGPJoCr+2sPY0r8hCYFQAYPr1U9EoCTYICORbJZMeucAo4v6AH9LxwDFx0\r\n8IpXkfwett+Q2AvMAQw6v0s0bqTJ20n4dLCfhdu3IdDgTXlg6My\/mGswao1f+BdE\r\ntdJ5iBL\/QMpowoz2SZeiYMtLOxf+NC5h2iVxd+ijZ0JjMEedSozz0y60QuVWaJ2J\r\nndSjEwhphcx6cGctnJ83w4CQkGurfGQKs0S5k+5zUxANulSufSiH9mC4n3rOEw2v\r\nqW3H6jg=\r\n=YW+U\r\n-----END PGP PUBLIC KEY BLOCK-----\r\n","env":"dev","notify":"http:\/\/www.baidu.com\/"}`
	buffer := bytes.NewBufferString(source)
	decoder := json.NewDecoder(buffer)
	var reqBody MultipleBody
	err := decoder.Decode(&reqBody)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	t.Log(reqBody)
	t.Log("PASS")
}
