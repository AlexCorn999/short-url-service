package apiserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/AlexCorn999/short-url-service/internal/app/gzip"
	"github.com/AlexCorn999/short-url-service/internal/app/logger"
	"github.com/AlexCorn999/short-url-service/internal/app/store"
	"github.com/go-chi/chi"
	bolt "go.etcd.io/bbolt"

	log "github.com/sirupsen/logrus"
)

// URL для JSON объекта
type shortenURL struct {
	URL string `json:"url"`
}

// URL для JSON объекта
type URLResult struct {
	ResultURL string `json:"result"`
}

// APIServer ...
type APIServer struct {
	storage *store.Db
	logger  *log.Logger
	config  *Config
	router  *chi.Mux
}

// New APIServer
func New(config *Config) *APIServer {
	db, err := bolt.Open(config.FilePath, 0666, nil)
	if err != nil {
		log.Fatal(err)
	}

	return &APIServer{
		config:  config,
		logger:  log.New(),
		router:  chi.NewRouter(),
		storage: store.NewDb(db),
	}
}

// Start APIServer
func (s *APIServer) Start() error {
	s.configureRouter()
	s.storage.CreateBacketURL()

	defer s.storage.Store.Close()

	if err := s.configureLogger(); err != nil {
		return err
	}

	s.logger.Info("starting api server")

	return http.ListenAndServe(s.config.bindAddr, s.router)
}

func (s *APIServer) configureRouter() {
	s.router = chi.NewRouter()

	s.router.Use(logger.WithLogging)
	s.router.Use(gzip.GzipHandle)
	s.router.Post("/api/shorten", s.ShortenURL)
	s.router.Post("/", s.StringAccept)
	s.router.Get("/{id}", s.StringBack)
	s.router.NotFound(badRequest)
}

func (s *APIServer) configureLogger() error {
	level, err := log.ParseLevel(s.config.LogLevel)

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

	url := store.NewURL(link, string(body))
	s.storage.WriteURL(url, idForData)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(link))
}

// StringBack принимает id и возвращает ссылку
func (s *APIServer) StringBack(w http.ResponseWriter, r *http.Request) {
	id := r.URL.String()

	var url store.URL

	s.storage.ReadURL(&url, id[1:])

	w.Header().Set("Location", url.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// ShortenURL принимает JSON-объект {"url":"<some_url>"}.
// Возвращает в ответ объект {"result":"<short_url>"}.
func (s *APIServer) ShortenURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var url shortenURL

	if err := json.Unmarshal(body, &url); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// проверка на пустую ссылку
	if len(strings.TrimSpace(string(url.URL))) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// запись в хранилище
	idForData := strconv.Itoa(store.IDStorage)
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

	urlNew := store.NewURL(link, url.URL)
	s.storage.WriteURL(urlNew, idForData)

	// запись ссылки в структуру ответа
	var result URLResult
	result.ResultURL = link

	objectJSON, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(objectJSON)
}
