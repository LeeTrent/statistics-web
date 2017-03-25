package main

import (
	"errors"
	"fmt"
	"github.com/LeeTrent/fileutil"
	"github.com/LeeTrent/statistics"
	"html/template"
	"net/http"
	"strconv"
)

const tmplDir string = "templates/*"
const userDir string = "./userfiles/"
const floatFormat string = "%.4f"
const debugging bool = true

//type CalcResults struct {
//	Data     []float64
//	Mean     float64
//	Variance float64
//	StdDev   float64
//}

type CalcResults struct {
	Mean string
	Variance string
	StdDev   string
}

var tmpl *template.Template

func init() {
	tmpl = template.Must(template.ParseGlob(tmplDir))
}

func main() {
	http.HandleFunc("/", index)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.ListenAndServe(":8080", nil)
}

func debug(val string) {
	fmt.Println(val)
}

func index(resp http.ResponseWriter, req *http.Request) {
	if debugging {
		fmt.Printf("req.Method: %+v\n", req.Method)
	}

	if req.Method == http.MethodGet {
		doIndexGet(resp, req)
	} else if req.Method == http.MethodPost {
		doIndexPost(resp, req)
	} else {
		doIndexBadRequest(resp, req)
	}
}

func doIndexGet(resp http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodGet {
		resp.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.ExecuteTemplate(resp, "index.gohtml", nil)
	}
}

func doIndexBadRequest(resp http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodGet {
		resp.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.ExecuteTemplate(resp, "index.gohtml", "Only GET and POST HTTP Methods are accepted.")
	}
}

func doIndexPost(resp http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodPost {

		fileName, err := writeUploadedFile(resp, req)
		if err != nil {
			fmt.Println(err)
			http.Error(resp, err.Error(), http.StatusInternalServerError)
			return
		}

		fileData, err := extractUploadedData(fileName)
		if err != nil {
			fmt.Println(err)
			http.Error(resp, err.Error(), http.StatusInternalServerError)
			return
		}

		mean := statistics.CalcMean(fileData)
		variance := statistics.CalcVarianceUsingMean(mean, fileData)
		stdDev := statistics.CalcStandardDeviationUsingVariance(variance)

		if debugging {
			fmt.Printf("sampleData: %+v\n", fileData)
			fmt.Printf("mean: %+v\n", mean)
			fmt.Printf("variance: %+v\n", variance)
			fmt.Printf("stdDev: %+v\n", stdDev)
		}

		//calcResults :=  CalcResults{Mean: mean, Variance: variance, StdDev: stdDev,}
		calcResults := CalcResults{
			Mean: fmt.Sprintf(floatFormat, mean),
			Variance: fmt.Sprintf(floatFormat, variance),
			StdDev: fmt.Sprintf(floatFormat, stdDev),
		}

		resp.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.ExecuteTemplate(resp, "index.gohtml", calcResults)
	}
}

func writeUploadedFile(resp http.ResponseWriter, req *http.Request) (string, error) {

	var fileName string

	if req.Method == http.MethodPost {

		// open uploaded file
		srcFile, srcHdr, err := req.FormFile("q")
		if err != nil {
			return fileName, err
		}
		defer srcFile.Close()

		fileName = srcHdr.Filename

		// write uploaded file to disk
		err = fileutil.WriteFileToDisk(srcFile, userDir, srcHdr.Filename)
		if err != nil {
			return fileName, err
		}

		return fileName, nil
	}

	// error condtion if not HTTP POST
	errMsg := fmt.Sprintf("HTTP POST method expected, instead got '%v'", req.Method)
	return fileName, errors.New(errMsg)
}

func extractUploadedData(fileName string) ([]float64, error) {

	var sampleData []float64

	fileData, err := fileutil.ReadFileFromDisk(userDir, fileName)
	if err != nil {
		return sampleData, err
	}

	for index, row := range fileData {
		if index >= 0 {
			val, err := strconv.ParseFloat(row[0], 64)
			if err == nil {
				sampleData = append(sampleData, val)
			} else {
				return sampleData, err
			}
		}
	}
	return sampleData, nil
}

// https://play.golang.org/p/Znf6wivRbI
