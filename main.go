package main

import (
	"flag"
	"log"
	"time"
)

var (
	bindAddrFlag         string
	storagePathFlag      string
	baseURLFlag          string
	cleaningIntervalFlag time.Duration
	lifeTimeFlag         time.Duration
)

func init() {
	flag.StringVar(&bindAddrFlag, "bind-addr", ":80", "Address to bind HTTP-server")
	flag.StringVar(&storagePathFlag, "storage-path", "./storage", "Storage base path")
	flag.StringVar(&baseURLFlag, "base-url", "http://localhost", "Base URL for report links")
	flag.DurationVar(&cleaningIntervalFlag, "cleaning-interval", 24*time.Hour, "Cleaning interval")
	flag.DurationVar(&lifeTimeFlag, "life-time", 7*24*time.Hour, "Life time for storage items")
}

func main() {
	flag.Parse()

	s, err := newStorage(storagePathFlag)
	if err != nil {
		log.Fatal(err)
	}

	srv := httpServer{
		s:       s,
		baseURL: baseURLFlag,
	}

	go srv.s.cleaning(cleaningIntervalFlag, lifeTimeFlag)

	err = srv.start(bindAddrFlag)
	if err != nil {
		log.Fatal(err)
	}
}
