package apiserver

import (
	"errors"
	"flag"
	"os"
	"strconv"
	"strings"
)

// Config ...
type Config struct {
	bindAddr     string
	ShortURLAddr string
	FilePath     string
	LogLevel     string
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{
		bindAddr: ":8080",
		FilePath: "/tmp/short-url-db.json",
		LogLevel: "debug",
	}
}

// NetAddress структура для проверки флага a
type NetAddress struct {
	Host string
	Port int
}

// URLAddress cтруктура для проверки флага b
type URLAddress struct {
	http string
	Host string
	Port int
}

// Имплементация интерфейса для флага b
func (u URLAddress) String() string {
	return u.http + ":" + u.Host + ":" + strconv.Itoa(u.Port)
}

func (u *URLAddress) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 3 {
		return errors.New("need address in a form host:port")
	}
	port, err := strconv.Atoi(hp[2])
	if err != nil {
		return err
	}
	u.http = hp[0]
	u.Host = hp[1]
	u.Port = port
	return nil
}

// Имплементация интерфейса для флага a
func (a NetAddress) String() string {
	return a.Host + ":" + strconv.Itoa(a.Port)
}

func (a *NetAddress) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return errors.New("need address in a form host:port")
	}
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		return err
	}
	a.Host = hp[0]
	a.Port = port
	return nil
}

// parseFlags обрабатывает аргументы командной строки и сохраняет их значения в соответствующих переменных
func (c *Config) ParseFlags() {

	addr := new(NetAddress)
	_ = flag.Value(addr)

	urlAddr := new(URLAddress)
	_ = flag.Value(urlAddr)

	flag.Var(addr, "a", "Net address host:port")
	flag.Var(urlAddr, "b", "address and port for short URL")
	filePath := flag.String("f", "/tmp/short-url-db.json", "filePath for URL")

	flag.Parse()
	c.FilePath = *filePath

	// проверка значения addr, чтобы записать в переменную bindAddr
	if addr.String() != ":0" {
		c.bindAddr = addr.String()
	}

	// проверка значения urlAddr, чтобы записать в переменную shortURLAddr
	if urlAddr.String() != "::0" {
		c.ShortURLAddr = urlAddr.String()
	}

	// Установка данных адреса запуска HTTP-сервера через переменные окружения
	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		c.bindAddr = envRunAddr
	}

	// Установка базового адреса результирующего сокращённого URL через переменные окружения
	if envShortURL := os.Getenv("BASE_URL"); envShortURL != "" {
		c.ShortURLAddr = envShortURL
	}

	// Установка пути к файлу сохранения URL через переменные окружения
	if envPath := os.Getenv("FILE_STORAGE_PATH"); envPath != "" {
		c.FilePath = envPath
	}
}
