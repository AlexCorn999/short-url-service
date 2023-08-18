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
	storage *store.DB
	logger  *log.Logger
	config  *Config
	router  *chi.Mux
}

// New APIServer
func New(config *Config) *APIServer {

	if len(strings.TrimSpace(config.FilePath)) != 0 {
		db, err := bolt.Open(config.FilePath, 0666, nil)
		if err != nil {
			log.Fatal(err)
		}
		return &APIServer{
			config:  config,
			logger:  log.New(),
			router:  chi.NewRouter(),
			storage: store.NewDB(db),
		}
	}

	return &APIServer{
		config:  config,
		logger:  log.New(),
		router:  chi.NewRouter(),
		storage: store.NewMemoryDB(),
	}
}

// Start APIServer
func (s *APIServer) Start() error {
	s.configureRouter()

	if len(strings.TrimSpace(s.config.databaseAddr)) != 0 {
		if err := s.configureStore(); err != nil {
			return err
		}
		defer s.storage.CloseDB()

	} else if len(strings.TrimSpace(s.config.FilePath)) != 0 {
		s.storage.CreateBacketURL()
		defer s.storage.Store.Close()
	}

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
	s.router.Get("/ping", s.Ping)
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

func (s *APIServer) configureStore() error {

	if len(strings.TrimSpace(s.config.databaseAddr)) == 0 {
		return nil
	}
	if err := s.storage.OpenDB(s.config.databaseAddr); err != nil {
		return err
	}
	if err := s.storage.InitTables(); err != nil {
		return err
	}
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
	store.NextID(&store.IDStorage)

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

	// разделение для записи БД / Файл / Память

	if len(strings.TrimSpace(s.config.databaseAddr)) != 0 {
		if err := s.storage.AddURL(url, idForData); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(link))
		return

	} else if len(strings.TrimSpace(s.config.FilePath)) != 0 {

		if err := s.storage.WriteURL(url, idForData); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(link))
		return

	} else {
		s.storage.MemoryDB[idForData] = string(body)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(link))
		return
	}

}

// StringBack принимает id и возвращает ссылку
func (s *APIServer) StringBack(w http.ResponseWriter, r *http.Request) {
	id := r.URL.String()

	var url store.URL

	if len(strings.TrimSpace(s.config.databaseAddr)) != 0 {
		addr, err := s.storage.AddrBack(id[1:])
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Location", addr)

	} else if len(strings.TrimSpace(s.config.FilePath)) != 0 {
		s.storage.ReadURL(&url, id[1:])
		w.Header().Set("Location", url.OriginalURL)

	} else {
		if _, ok := s.storage.MemoryDB[id[1:]]; !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		link := s.storage.MemoryDB[id[1:]]
		w.Header().Set("Location", link)
	}

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
	store.NextID(&store.IDStorage)

	hostForLink := r.Host
	var link string

	// проверка для работы флага b
	if s.config.ShortURLAddr != "" {
		hostForLink = s.config.ShortURLAddr
		link = fmt.Sprintf("%s/%s", hostForLink, idForData)
	} else {
		link = fmt.Sprintf("http://%s/%s", hostForLink, idForData)
	}

	if len(strings.TrimSpace(s.config.databaseAddr)) != 0 {
		urlNew := store.NewURL(link, url.URL)
		if err := s.storage.AddURL(urlNew, idForData); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

	} else if len(strings.TrimSpace(s.config.FilePath)) != 0 {
		urlNew := store.NewURL(link, url.URL)
		if err := s.storage.WriteURL(urlNew, idForData); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		s.storage.MemoryDB[idForData] = url.URL
	}

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

// Ping проверяет соединение с базой данных.
func (s *APIServer) Ping(w http.ResponseWriter, r *http.Request) {
	// проверка на работу только с базой данных.
	if len(strings.TrimSpace(s.config.databaseAddr)) != 0 {
		if err := s.storage.CheckPing(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}
