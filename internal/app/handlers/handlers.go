package handlers

import (
	"fmt"
	"net/http"
	"strconv"
)

var (
	IDStore = 1
)

type Storage struct {
	data map[string]string
}

func NewStorage() *Storage {
	return &Storage{data: make(map[string]string)}
}

// StringAccept принимает ссылку и возвращает закодированную ссылку
func (s *Storage) StringAcceptAndBack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if r.Method == http.MethodPost {
		// парсинг тела запроса POST
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		body := ""
		for k := range r.Form {
			body += k
		}

		// запись в хранилище
		idForData := strconv.Itoa(IDStore)
		s.data[idForData] = body
		IDStore++

		link := fmt.Sprintf("http:%s/%s", r.Host, idForData)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(link))
		return
	}

	if r.Method == http.MethodGet {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		body := r.URL.String()

		if _, ok := s.data[body[1:]]; !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		link := s.data[body[1:]]

		//link := fmt.Sprintf("http:%s", r.Host)
		w.Header().Set("Location", link)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
}

/*
// StringBack возвращает ссылку по id
func StringBack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body := r.URL.String()

	link := fmt.Sprintf("http:%s%s", r.Host, body)
	w.Header().Set("Location", link)
	w.WriteHeader(http.StatusTemporaryRedirect)
}*/
