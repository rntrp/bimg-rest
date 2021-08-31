package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/h2non/bimg"
)

const maxMemory int64 = 1024 * 64
const maxFileSize int64 = 1024 * 1024 * 512

func Welcome(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("welcome"))
}

func Scale(w http.ResponseWriter, r *http.Request) {
	printMemUsage()
	query := r.URL.Query()
	width, _ := strconv.Atoi(query.Get("width"))
	height, _ := strconv.Atoi(query.Get("height"))
	format := getFormat(query.Get("format"))
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	f, fh, err := r.FormFile("image")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("File 'image' could not be processed."))
		return
	}
	defer f.Close()
	if fh.Size > maxFileSize {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		fmt.Fprintf(w, "Max file size is %d", maxFileSize)
		return
	}
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(buf) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("File 'image' is empty."))
		return
	}
	options := getOptions(width, height, format)
	o, err := bimg.NewImage(buf).Process(options)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(o)
	printMemUsage()
}

func getFormat(format string) bimg.ImageType {
	switch strings.ToLower(format) {
	case "png":
		return bimg.PNG
	case "webp":
		return bimg.WEBP
	case "gif":
		return bimg.GIF
	case "heif", "heic":
		return bimg.HEIF
	case "avif":
		return bimg.AVIF
	default:
		return bimg.JPEG
	}
}

func getOptions(width, height int, format bimg.ImageType) bimg.Options {
	return bimg.Options{
		Width:         width,
		Height:        height,
		Type:          format,
		Compression:   9,
		Quality:       99,
		Extend:        bimg.ExtendWhite,
		Background:    bimg.Color{R: 0xFF, G: 0xFF, B: 0xFF},
		Embed:         true,
		Enlarge:       true,
		StripMetadata: true,
		NoProfile:     true,
	}
}

func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	bToMiB := func(b uint64) string {
		return strconv.FormatFloat(float64(b)/1_048_576, 'f', 3, 64)
	}
	fmt.Printf("Alloc = %v MiB"+
		"\tTotalAlloc = %v MiB"+
		"\tSys = %v MiB"+
		"\tNumGC = %v\n",
		bToMiB(m.Alloc),
		bToMiB(m.TotalAlloc),
		bToMiB(m.Sys),
		m.NumGC)
}
