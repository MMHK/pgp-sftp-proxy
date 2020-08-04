package lib

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/service/textract"
	"os"
	"testing"
)

func Test_OCRService_GetPDFText(t *testing.T) {
	conf := &AWSOption{
		Region: "ap-southeast-1",
		Bucket: "s3.test.mixmedia.com",
		AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}

	//t.Logf("%+v", conf)

	ocr, err := NewOCRService(conf)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	pdf, err := os.Open(getLocalPath("../test/temp/dahsing-CI-sample.pdf"))
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	defer pdf.Close()


	err = ocr.GetPDFText(pdf)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
}

func Test_OCRService_GetFormData(t *testing.T) {
	conf := &AWSOption{
		Region: "ap-southeast-1",
		Bucket: "s3.test.mixmedia.com",
		AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}

	//t.Logf("%+v", conf)

	ocr, err := NewOCRService(conf)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	pdf, err := os.Open(getLocalPath("../test/temp/dahsing-SCHEDULE-sample.pdf"))
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	defer pdf.Close()


	mapping, err := ocr.GetFormData(pdf)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	t.Log(mapping)
}

func Test_OCRService_GetFormDataFromFile(t *testing.T) {
	conf := &AWSOption{
		Region: "ap-southeast-1",
		Bucket: "s3.test.mixmedia.com",
		AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}

	//t.Logf("%+v", conf)

	ocr, err := NewOCRService(conf)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	mapping, err := ocr.GetFormDataFromFile(getLocalPath("../test/temp/dahsing-SCHEDULE-sample.pdf"))
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	t.Log(mapping)
}

func Test_GetDocumentAnalysisOutput(t *testing.T) {
	resp, err := os.Open(getLocalPath("../test/temp/apiResponse.json"))
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	defer resp.Close()

	outPut := new(textract.AnalyzeDocumentOutput)

	decoder := json.NewDecoder(resp)
	err = decoder.Decode(outPut)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	//t.Logf("%+v", outPut)

	words := make(map[string]*string, 0)
	formList := make([]*KeyValueRaw, 0)
	values := make(map[string][]*string, 0)
	for _, block := range outPut.Blocks {
		blockType := *block.BlockType
		if blockType == textract.BlockTypeWord {
			words[*block.Id] = block.Text
		}

		if blockType == textract.BlockTypeKeyValueSet &&
			len(block.EntityTypes) > 0 &&
			(*block.EntityTypes[0]) == textract.EntityTypeValue {
			valueChild := make([]*string, 0)

			for _, item := range block.Relationships {
				if (*item.Type) == textract.RelationshipTypeChild {
					valueChild = append(valueChild, item.Ids...)
				}
			}

			values[*block.Id] = valueChild
		}

		if blockType == textract.BlockTypeKeyValueSet &&
			len(block.EntityTypes) > 0 &&
			(*block.EntityTypes[0]) == textract.EntityTypeKey {

			keyList := make([]*string, 0)
			valueList := make([]*string, 0)

			for _, item := range block.Relationships {
				if (*item.Type) == textract.RelationshipTypeValue {
					valueList = append(valueList, item.Ids...)
				}
				if (*item.Type) == textract.RelationshipTypeChild {
					keyList = append(keyList, item.Ids...)
				}
			}

			formList = append(formList, &KeyValueRaw{
				Key:   keyList,
				Value: valueList,
			})
		}
	}

	mapping := make(map[string]string, 0)

	for _, kv := range formList {
		mapping[kv.GetKey(words)] = kv.GetValue(words, values)
	}

	for k, v := range mapping {
		t.Logf("%s\t=>\t%s", k, v)
	}

}