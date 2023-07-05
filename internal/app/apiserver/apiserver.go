package apiserver

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/AlexCorn999/short-url-service/internal/app/store"
	"github.com/go-chi/chi"
)

// APIServer ...
type APIServer struct {
	storage store.Storage
	config  *Config
	router  *chi.Mux
}

// Start ...
func (s *APIServer) Start() error {
	s.config = NewConfig()
	s.config.parseFlags()

	s.configureRouter()
	s.storage = *store.NewStorage()
	return http.ListenAndServe(s.config.bindAddr, s.router)
}

func (s *APIServer) configureRouter() {
	s.router = chi.NewRouter()

	s.router.Post("/", s.StringAccept)
	s.router.Get("/{id}", s.StringBack)
	s.router.NotFound(notFoundError)
}

func notFoundError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

// StringAccept принимает ссылку и возвращает закодированную ссылку
func (s *APIServer) StringAccept(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// запись в хранилище
	idForData := strconv.Itoa(store.IDStorage)
	s.storage.Data[idForData] = string(body)
	store.IDStorage++

	link := fmt.Sprintf("http://%s/%s", s.config.shortURLAddr, idForData)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(link))
}

// StringBack принимает id и возвращает ссылку
func (s *APIServer) StringBack(w http.ResponseWriter, r *http.Request) {
	id := r.URL.String()

	if _, ok := s.storage.Data[id[1:]]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	link := s.storage.Data[id[1:]]

	w.Header().Set("Location", link)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
