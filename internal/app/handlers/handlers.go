package handlers

import (
	"fmt"
	"io"
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
	// обработка POST метода
	if r.Method == http.MethodPost {

		bodyPost := r.URL.String()
		if bodyPost != "/" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		/*
			// парсинг тела запроса POST
			if err := r.ParseForm(); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			body := ""
			for k := range r.Form {
				body += k
			}*/

		// запись в хранилище
		idForData := strconv.Itoa(IDStore)
		s.data[idForData] = string(body)
		IDStore++

		link := fmt.Sprintf("http://%s/%s", r.Host, idForData)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(link))
		return
	}

	// обработка GET метода
	if r.Method == http.MethodGet {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		id := r.URL.String()

		if _, ok := s.data[id[1:]]; !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		link := s.data[id[1:]]

		w.Header().Set("Location", link)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
}
