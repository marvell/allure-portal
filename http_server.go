package main

import (
	"archive/zip"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

const maxUploadingFileSize = 128 << 20 // 128 Mb
const multipartFileName = "file"

type httpServer struct {
	s       *storage
	baseURL string
}

func (s *httpServer) getRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/upload", s.uploadHandler)
	mux.Handle("/", http.FileServer(http.Dir(s.s.basePath)))

	return mux
}

func (s *httpServer) start(bindAddr string) error {
	srv := http.Server{
		Addr:    bindAddr,
		Handler: s.getRouter(),
	}

	log.Printf("http-server is staring on %s", bindAddr)

	return srv.ListenAndServe()
}

func (s *httpServer) uploadHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		s.writeError(res, http.StatusNotFound, "wrong method")
		return
	}

	// validating file size
	req.Body = http.MaxBytesReader(res, req.Body, maxUploadingFileSize)
	err := req.ParseMultipartForm(maxUploadingFileSize)
	if err != nil {
		s.writeError(res, http.StatusBadRequest, err.Error())
		return
	}

	group := strings.TrimSpace(req.FormValue("group"))
	if group == "" {
		s.writeError(res, http.StatusBadRequest, "'group' param should not be empty")
		return
	}

	project := strings.TrimSpace(req.FormValue("project"))
	if project == "" {
		s.writeError(res, http.StatusBadRequest, "'project' param should not be empty")
		return
	}

	versionKey := strings.TrimSpace(req.FormValue("version"))
	if versionKey == "" {
		s.writeError(res, http.StatusBadRequest, "'version' param should not be empty")
		return
	}

	versionTS := time.Now()
	if ts := strings.TrimSpace(req.FormValue("ts")); ts != "" {
		versionTS, err = time.Parse(time.RFC3339, ts)
		if err != nil {
			return
		}
	}

	sk := newStorageKey(group, project, newStorageKeyVersion(versionKey, versionTS))

	log.Printf("uploading file (size: %d) to %s", req.ContentLength, sk.getPath())

	// obtaining file data
	file, _, err := req.FormFile(multipartFileName)
	if err != nil {
		s.writeError(res, http.StatusBadRequest, err.Error())
		return
	}
	defer file.Close()

	// TODO: check file (non empty)

	r, err := zip.NewReader(file, req.ContentLength)
	if err != nil {
		s.writeError(res, http.StatusInternalServerError, err.Error())
		return
	}

	err = s.s.putResults(r, sk)
	if err != nil {
		s.writeError(res, http.StatusInternalServerError, err.Error())
		return
	}

	err = s.s.copyHistory(sk)
	if err != nil {
		s.writeError(res, http.StatusInternalServerError, err.Error())
		return
	}

	err = s.s.generateReport(sk)
	if err != nil {
		s.writeError(res, http.StatusInternalServerError, err.Error())
		return
	}

	err = s.s.createLastVersionSymlink(sk)
	if err != nil {
		s.writeError(res, http.StatusInternalServerError, err.Error())
		return
	}

	fmt.Fprintf(res, "%s/%s/report/\n", s.baseURL, sk.getPath())
}

func (s *httpServer) writeError(res http.ResponseWriter, code int, msg string) {
	res.WriteHeader(code)
	_, err := res.Write([]byte(msg))
	if err != nil {
		log.Printf("ERROR: %s", err)
	}

	if code >= 500 {
		log.Printf("ERROR: code=%d %s", code, msg)
	}
}
