package apiserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/AlexCorn999/short-url-service/internal/app/auth"
	"github.com/AlexCorn999/short-url-service/internal/app/filestorage"
	"github.com/AlexCorn999/short-url-service/internal/app/gzip"
	"github.com/AlexCorn999/short-url-service/internal/app/logger"
	"github.com/AlexCorn999/short-url-service/internal/app/memorystorage"
	"github.com/AlexCorn999/short-url-service/internal/app/store"
	"github.com/go-chi/chi"

	log "github.com/sirupsen/logrus"
)

var (
	authForFlag = false
	authString  string
	tknStr      string
)

type batchURL struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
	shortURL      string
}

type resultBatchURL struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

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
	store.Database
	initialized bool
	typeStore   string
	logger      *log.Logger
	config      *Config
	router      *chi.Mux
}

// New APIServer
func New(config *Config) *APIServer {
	return &APIServer{
		config:      config,
		initialized: false,
		logger:      log.New(),
		router:      chi.NewRouter(),
	}
}

// Start APIServer
func (s *APIServer) Start() error {
	s.configureRouter()

	if err := s.configureLogger(); err != nil {
		return err
	}

	if err := s.configureStore(); err != nil {
		return err
	}

	if s.typeStore == "database" {
		defer s.Database.Close()
	} else if s.typeStore == "file" {
		defer s.Database.Close()
	}

	s.logger.Info("starting api server")

	return http.ListenAndServe(s.config.bindAddr, s.router)
}

func (s *APIServer) configureRouter() {
	s.router = chi.NewRouter()
	s.router.Use(s.Auth)
	s.router.Use(logger.WithLogging)
	s.router.Use(gzip.GzipHandle)
	s.router.Post("/api/shorten/batch", s.BatchURL)
	s.router.Post("/api/shorten", s.ShortenURL)
	s.router.Post("/", s.StringAccept)
	s.router.Get("/{id}", s.StringBack)
	s.router.Get("/ping", s.Ping)
	s.router.Get("/api/user/urls", s.GetAllURL)
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

	if len(strings.TrimSpace(s.config.databaseAddr)) != 0 {

		db, err := store.NewPostgres(s.config.databaseAddr)
		if err != nil {
			return err
		}
		s.Database = db
		s.typeStore = "database"
		// переписываем значение из базы для user_id
		auth.ID, err = s.Database.InitID()
		auth.ID++
		if err != nil {
			return err
		}

	} else if len(strings.TrimSpace(s.config.FilePath)) != 0 {
		db, err := filestorage.NewBoltDB(s.config.FilePath)
		if err != nil {
			return err
		}
		s.Database = db
		s.typeStore = "file"
	} else {
		db := memorystorage.NewMemoryStorage()
		s.Database = db
		s.typeStore = "local"
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

	// для авторизации
	if !authForFlag {
		c, err := r.Cookie("token")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		tknStr = c.Value
	} else {
		tknStr = authString
	}

	creator, err := auth.GetUserID(tknStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s.logger.Info("Юзер ", creator, " ссылка ", string(body))

	url := store.NewURL(link, string(body), creator)
	if err = s.Database.WriteURL(url, creator, &idForData); err != nil {
		// проверка, что ссылка уже есть в базе
		if errors.Is(err, store.ErrConfilict) {

			res, err := s.Database.Conflict(url)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(res))
			return
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	// проверка для работы флага b БД
	if s.typeStore == "database" {
		if s.config.ShortURLAddr != "" {
			hostForLink = s.config.ShortURLAddr
			link = fmt.Sprintf("%s/%s", hostForLink, idForData)
		} else {
			link = fmt.Sprintf("http://%s/%s", hostForLink, idForData)
		}
		urlResult := store.NewURL(link, string(body), creator)
		// тут нужно перезаписать значения в базе
		if err := s.Database.RewriteURL(urlResult); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(link))

}

// StringBack принимает id и возвращает ссылку
func (s *APIServer) StringBack(w http.ResponseWriter, r *http.Request) {
	id := r.URL.String()

	var url store.URL

	if err := s.Database.ReadURL(&url, id[1:]); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
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

	/// для авторизации
	var tknStr string
	if !authForFlag {
		c, err := r.Cookie("token")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		tknStr = c.Value
	} else {
		tknStr = authString
	}

	creator, err := auth.GetUserID(tknStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	urlNew := store.NewURL(link, url.URL, creator)
	if err := s.Database.WriteURL(urlNew, creator, &idForData); err != nil {
		// проверка, что ссылка уже есть в базе
		if errors.Is(err, store.ErrConfilict) {

			var result shortenURL
			objectJSON, err := json.Marshal(result)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			w.Write(objectJSON)
			return

		} else {
			fmt.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	// проверка для работы флага b БД
	if s.typeStore == "database" {
		if s.config.ShortURLAddr != "" {
			hostForLink = s.config.ShortURLAddr
			link = fmt.Sprintf("%s/%s", hostForLink, idForData)
		} else {
			link = fmt.Sprintf("http://%s/%s", hostForLink, idForData)
		}
		urlResult := store.NewURL(link, url.URL, creator)
		// тут нужно перезаписать значения в базе
		if err := s.Database.RewriteURL(urlResult); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
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

// BatchURL принимает множество URL в формате JSON.
// Возвращает в ответ множество объектов JSON.
func (s *APIServer) BatchURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var urls []batchURL

	if err := json.Unmarshal(body, &urls); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// проверка на пустую ссылку
	for _, url := range urls {
		if len(strings.TrimSpace(url.OriginalURL)) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	for i := 0; i < len(urls); i++ {
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

		// для авторизации
		if !authForFlag {
			c, err := r.Cookie("token")
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			tknStr = c.Value
		} else {
			tknStr = authString
		}

		creator, err := auth.GetUserID(tknStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		urlNew := store.NewURL(link, urls[i].OriginalURL, creator)
		if err := s.Database.WriteURL(urlNew, creator, &idForData); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// проверка для работы флага b БД
		if s.typeStore == "database" {
			if s.config.ShortURLAddr != "" {
				hostForLink = s.config.ShortURLAddr
				link = fmt.Sprintf("%s/%s", hostForLink, idForData)
			} else {
				link = fmt.Sprintf("http://%s/%s", hostForLink, idForData)
			}
			urlResult := store.NewURL(link, urls[i].OriginalURL, creator)
			// тут нужно перезаписать значения в базе
			if err := s.Database.RewriteURL(urlResult); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		urls[i].shortURL = link
	}

	// запись ссылки в структуру ответа
	var result []resultBatchURL

	for _, url := range urls {
		res := resultBatchURL{
			CorrelationID: url.CorrelationID,
			ShortURL:      url.shortURL,
		}
		result = append(result, res)
	}

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
	if s.typeStore == "database" {
		if err := s.Database.CheckPing(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

// GetAllURL возвращает пользователю все сокращенные им url.
func (s *APIServer) GetAllURL(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("token")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	tknStr = c.Value

	creator, err := auth.GetUserID(tknStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s.logger.Info("Узнаем у юзера ", creator)
	result, err := s.Database.GetAllURL(creator)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	type resultURL struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}

	resultForJSON := make([]resultURL, len(result))

	if s.typeStore == "database" {
		for i := 0; i < len(result); i++ {
			resultForJSON[i].OriginalURL = result[i].ShortURL
			resultForJSON[i].ShortURL = result[i].OriginalURL
		}
	} else {
		for i := 0; i < len(result); i++ {
			resultForJSON[i].OriginalURL = result[i].OriginalURL
			resultForJSON[i].ShortURL = result[i].ShortURL
		}
	}

	if len(resultForJSON) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	objectJSON, err := json.Marshal(resultForJSON)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(objectJSON)
}

func (s *APIServer) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenFlag := false
		var token string

		// проверка на cуществование cookie
		c, err := r.Cookie("token")
		if err != nil {
			if err == http.ErrNoCookie {
				token, err = auth.BuildJWTString()
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				tokenFlag = true
			} else {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

		}

		if !tokenFlag {
			tknStr = c.Value
		} else {
			tknStr = token
		}

		if tokenFlag {
			authForFlag = true
			authString = token
			http.SetCookie(w, &http.Cookie{
				Name:  "token",
				Value: token,
			})

		}

		next.ServeHTTP(w, r)
	})
}
