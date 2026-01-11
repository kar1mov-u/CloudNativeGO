package main

import (
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

// var store = make(map[string]string)

type Store struct {
	storage map[string]string
	sync.RWMutex
}

var logger TransactionLogger
var s = Store{storage: map[string]string{}}

func (s *Store) Put(key, value string) error {
	s.Lock()
	s.storage[key] = value
	s.Unlock()
	return nil
}

var ErrorNoSuchKey = errors.New("no such key")

func (s *Store) Get(key string) (string, error) {
	s.RLock()
	value, ok := s.storage[key]
	s.RUnlock()

	if !ok {
		return "", ErrorNoSuchKey
	}
	return value, nil
}

func (s *Store) Delete(key string) error {
	s.Lock()
	delete(s.storage, key)
	s.Unlock()
	return nil
}

func initializeLogger() error {
	var err error
	logger, err = NewFileTransactionLogger("transaction.log")
	if err != nil {
		return err
	}

	events, errors := logger.ReadEvents()

	e := Event{}
	ok := true

	for ok && err == nil {
		select {
		case err, ok = <-errors:
		case e, ok = <-events:
			switch e.EventType {
			case EventDelete:
				err = s.Delete(e.Key)
			case EventPut:
				err = s.Put(e.Key, e.Value)
			}
		}
	}
	logger.Run()
	return err
}

func main() {
	initializeLogger()
	r := mux.NewRouter()

	r.HandleFunc("/v1/key/{key}", putHandler).Methods("PUT")
	r.HandleFunc("/v1/key/{key}", readHandler).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", r))
}
