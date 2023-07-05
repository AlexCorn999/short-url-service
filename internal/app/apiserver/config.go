package apiserver

import "flag"

type Config struct {
	bindAddr     string
	shortURLAddr string
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{}
}

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func (c *Config) parseFlags() {

	// регистрируем переменную flagRunAddr
	// как аргумент -a со значением :8080 по умолчанию
	flag.StringVar(&c.bindAddr, "a", ":8080", "address and port to run server")
	// регистрируем переменную flagRunAddr
	// как аргумент -a со значением :8080 по умолчанию
	flag.StringVar(&c.shortURLAddr, "b", "localhost:8080", "address and port for short URL")
	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()
}
