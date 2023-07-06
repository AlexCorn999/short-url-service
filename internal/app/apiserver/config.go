package apiserver

import (
	//"errors"
	"flag"
	//"strconv"
	//"strings"
)

type Config struct {
	bindAddr     string
	shortURLAddr string
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{}
}

type NetAddress struct {
	Host string
	Port int
}

/*
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
}*/

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func (c *Config) parseFlags() {
	// адрес по умолчанию
	c.bindAddr = "localhost:8080"
	//addr := new(NetAddress)
	//_ = flag.Value(addr)

	//flag.Var(addr, "a", "Net address host:port")
	flag.StringVar(&c.shortURLAddr, "b", "localhost:8080", "address and port for short URL")
	// парсим переданные серверу аргументы в зарегистрированные переменные
	//flag.Parse()

	//c.bindAddr = addr.String()
}
