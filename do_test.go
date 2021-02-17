package main

import (
	"os"
	"testing"
	"time"
)

type FakeFile struct {
	name     string
	contents string
	mode     os.FileMode
	offset   int
}

func (f *FakeFile) Name() string {
	// A bit of a cheat: we only have a basename, so that's also ok for FileInfo.
	return f.name
}

func (f *FakeFile) Size() int64 {
	return int64(len(f.contents))
}

func (f *FakeFile) Mode() os.FileMode {
	return f.mode
}

func (f *FakeFile) ModTime() time.Time {
	return time.Time{}
}

func (f *FakeFile) IsDir() bool {
	return false
}

func (f *FakeFile) Sys() interface{} {
	return nil
}

func TestParseArcDate(t *testing.T) {
	arcName := "untitled_26-10-2019_15.34.53.txt"
	confName := "untitled.txt"
	expectedResult := time.Date(2019, 10, 26, 15, 34, 53, 0, time.Local)

	result, err := ParseArcDate(arcName, confName)
	if result != expectedResult {
		t.Errorf("(arcName: \"%s\", confName: \"%s\")\ttime is different from expected, result: %s",
			arcName, confName, result)
	}
	if err != nil {
		t.Errorf("(arcName: \"%s\", confName: \"%s\")\terror is different from expected, err: %s",
			arcName, confName, err)
	}
}

func TestParseArcDate2(t *testing.T) {
	arcName := "._untitled_25-02-2019_23.00.36.txt"
	confName := "untitled.txt"
	expectedResult := time.Date(0001, 01, 01, 00, 00, 00, 0, time.UTC)
	expectedError := "the length of the file name \"" + arcName + "\" differs from the required length"

	result, err := ParseArcDate(arcName, confName)
	if err == nil || err.Error() != expectedError {
		t.Errorf("(arcName: \"%s\", confName: \"%s\")\terror is different from expected, err: %s",
			arcName, confName, err)
	}
	if result != expectedResult {
		t.Errorf("(arcName: \"%s\", confName: \"%s\")\ttime is different from expected, result: %s",
			arcName, confName, result)
	}
}

//Testing func ParseArcDate. Filename from webdav archive differs from the required length. Expected Error
func TestParseArcDate3(t *testing.T) {
	arcName := "untitled.txt"
	confName := "untitled.txt"
	expectedResult := time.Date(0001, 01, 01, 00, 00, 00, 0, time.UTC)
	expectedError := "the length of the file name \"" + arcName + "\" differs from the required length"

	result, err := ParseArcDate(arcName, confName)
	if err == nil || err.Error() != expectedError {
		t.Errorf("(arcName: \"%s\", confName: \"%s\")\terror is different from expected, err: %s",
			arcName, confName, err)
	}
	if result != expectedResult {
		t.Errorf("(arcName: \"%s\", confName: \"%s\")\ttime is different from expected, result: %s",
			arcName, confName, result)
	}
}

//Testing func ParseArcDate. Filename from webdav archive differs from the required length. Expected Error
func TestParseArcDate4(t *testing.T) {
	arcName := "._untitled.txt"
	confName := "untitled.txt"
	expectedResult := time.Date(0001, 01, 01, 00, 00, 00, 0, time.UTC)
	expectedError := "the length of the file name \"" + arcName + "\" differs from the required length"

	result, err := ParseArcDate(arcName, confName)
	if err == nil || err.Error() != expectedError {
		t.Errorf("(arcName: \"%s\", confName: \"%s\")\terror is different from expected, err: %s",
			arcName, confName, err)
	}
	if result != expectedResult {
		t.Errorf("(arcName: \"%s\", confName: \"%s\")\ttime is different from expected, result: %s",
			arcName, confName, result)
	}
}

//Testing func ParseArcDate. Filename from webdav archive contains offset
func TestParseArcDate5(t *testing.T) {
	arcName := "untitled_26-10-2019_15.34.53+0300.txt"
	confName := "untitled.txt"
	expectedResult := time.Date(2019, 10, 26, 15, 34, 53, 0,
		time.FixedZone("", 3*60*60))

	result, err := ParseArcDate(arcName, confName)
	if err != nil {
		t.Errorf("(arcName: \"%s\", confName: \"%s\")\terror is different from expectedResult, err: %s",
			arcName, confName, err)
		return
	}
	_, resultOffset := result.Zone()
	_, expectedOffset := expectedResult.Zone()
	//log.Printf("%+v", expectedResult)
	//log.Printf("%+v", result)
	if result.Nanosecond() != expectedResult.Nanosecond() || resultOffset != expectedOffset {
		t.Errorf("(arcName: \"%s\", confName: \"%s\")\ttime is different from expectedResult, result: %s",
			arcName, confName, result)
	}
}

func TestDoGetArcLastDate(t *testing.T) {
	//var filesArch []os.FileInfo
	var backupFileName string
	var result time.Time
	var err error

	var testFile = &FakeFile{
		name:     "untitled_26-10-2019_18.34.53+0300.txt",
		contents: "Testing, Jim", // 13 bytes, another odd number.
		mode:     0644,
	}
	backupFileName = "untitled.txt"
	filesArch := make([]os.FileInfo, 0)
	filesArch = append(filesArch, testFile)
	result, err = DoGetArcLastDate(backupFileName, filesArch)
	expected := time.Date(2019, 10, 26, 18, 34, 53, 0,
		time.FixedZone("", 3*60*60))
	_, resultOffset := result.Zone()
	_, expectedOffset := expected.Zone()
	if result.Nanosecond() != expected.Nanosecond() || resultOffset != expectedOffset {
		t.Errorf("(arcName: \"%s\", backupFileName: \"%s\")\ttime is different from expected, result: %s, "+
			"expected: %s", testFile.name, backupFileName, result, expected)
	}
	if err != nil {
		t.Errorf("(arcName: \"%s\", backupFileName: \"%s\")\terror is different from expected, err: %s",
			testFile.name, backupFileName, err)
	}
}
