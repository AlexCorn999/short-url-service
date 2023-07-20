package apiserver

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/AlexCorn999/short-url-service/internal/app/store"
	"github.com/go-chi/chi"

	"github.com/sirupsen/logrus"
)

// APIServer ...
type APIServer struct {
	storage store.Storage
	logger  *logrus.Logger
	config  *Config
	router  *chi.Mux
}

// New APIServer
func New(config *Config) *APIServer {
	return &APIServer{
		config:  config,
		logger:  logrus.New(),
		router:  chi.NewRouter(),
		storage: *store.NewStorage(),
	}
}

// Start APIServer
func (s *APIServer) Start() error {
	s.configureRouter()

	if err := s.configureStore(); err != nil {
		return err
	}

	if err := s.configureLogger(); err != nil {
		return err
	}

	s.logger.Info("starting api server")

	return http.ListenAndServe(s.config.bindAddr, s.router)
}

func (s *APIServer) configureRouter() {
	s.router = chi.NewRouter()
	s.router.Post("/", s.StringAccept)
	s.router.Get("/{id}", s.StringBack)
	s.router.NotFound(badRequest)
}

func (s *APIServer) configureStore() error {
	st := store.NewStorage()

	s.storage = *st

	return nil
}

func (s *APIServer) configureLogger() error {
	level, err := logrus.ParseLevel(s.config.LogLevel)

	if err != nil {
		return err
	}

	s.logger.SetLevel(level)
	return nil
}

// badRequest задает ошибку 400 по умолчанию на неизвестные запросы
func badRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

// StringAccept принимает ссылку и возвращает закодированную ссылку
func (s *APIServer) StringAccept(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// пероверка на пустую ссылку
	if len(strings.TrimSpace(string(body))) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// запись в хранилище
	idForData := strconv.Itoa(store.IDStorage)
	s.storage.Data[idForData] = string(body)
	store.IDStorage++

	hostForLink := r.Host
	var link string

	// проверка для работы флага b
	if s.config.ShortURLAddr != "" {
		hostForLink = s.config.ShortURLAddr
		link = fmt.Sprintf("%s/%s", hostForLink, idForData)
	} else {
		link = fmt.Sprintf("http://%s/%s", hostForLink, idForData)
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(link))
}

// StringBack принимает id и возвращает ссылку
func (s *APIServer) StringBack(w http.ResponseWriter, r *http.Request) {
	id := r.URL.String()

	if _, ok := s.storage.Data[id[1:]]; !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	link := s.storage.Data[id[1:]]

	w.Header().Set("Location", link)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
