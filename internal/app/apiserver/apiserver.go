package apiserver

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AlexCorn999/short-url-service/internal/app/store"
	"github.com/go-chi/chi"

	log "github.com/sirupsen/logrus"
)

type (
	// Структура для хранения сведений об ответе для middleware
	responseData struct {
		status int
		size   int
	}

	// Реализация http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// APIServer ...
type APIServer struct {
	storage store.Storage
	logger  *log.Logger
	config  *Config
	router  *chi.Mux
}

// New APIServer
func New(config *Config) *APIServer {
	return &APIServer{
		config:  config,
		logger:  log.New(),
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

	s.router.Use(WithLogging)
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
	level, err := log.ParseLevel(s.config.LogLevel)

	if err != nil {
		return err
	}

	s.logger.SetLevel(level)
	return nil
}

// WithLogging выполняет функцию middleware с логированием.
// Содержит сведения о URI, методе запроса и времени, затраченного на его выполнение.
// Сведения об ответах должны содержать код статуса и размер содержимого ответа.
func WithLogging(next http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		next.ServeHTTP(&lw, r)

		duration := time.Since(start)

		log.WithFields(log.Fields{
			"uri":      r.RequestURI,
			"method":   r.Method,
			"duration": duration,
			"status":   responseData.status,
			"size":     responseData.size,
		}).Info("request details: ")
	}
	return http.HandlerFunc(logFn)
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
