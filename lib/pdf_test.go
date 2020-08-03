package lib

import "testing"

func Test_GetLines(t *testing.T) {

	file := getLocalPath("../test/temp/dahsing-CI-sample.pdf")

	rows, err := GetLines(file)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	t.Logf("%+v", rows)
}