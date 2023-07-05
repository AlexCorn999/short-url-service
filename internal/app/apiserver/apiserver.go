package apiserver

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/AlexCorn999/short-url-service/internal/app/store"
)

const addr = ":8080"

// APIServer ...
type APIServer struct {
	storage store.Storage
	router  *http.ServeMux
}

// Start ...
func (s *APIServer) Start() error {
	s.configureRouter()
	s.storage = *store.NewStorage()
	return http.ListenAndServe(addr, s.router)
}

func (s *APIServer) configureRouter() {
	s.router = http.NewServeMux()
	s.router.HandleFunc("/", s.StringAcceptAndBack)
}

// StringAccept принимает ссылку и возвращает закодированную ссылку
func (s *APIServer) StringAcceptAndBack(w http.ResponseWriter, r *http.Request) {
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

		// запись в хранилище
		idForData := strconv.Itoa(store.IDStorage)
		s.storage.Data[idForData] = string(body)
		//	s.data[idForData] = string(body)
		store.IDStorage++

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

		if _, ok := s.storage.Data[id[1:]]; !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		link := s.storage.Data[id[1:]]

		w.Header().Set("Location", link)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
}
