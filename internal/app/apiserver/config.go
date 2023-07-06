package apiserver

import (
	"errors"
	"flag"
	"strconv"
	"strings"
)

type Config struct {
	bindAddr     string
	ShortURLAddr string
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{
		bindAddr: ":8080",
	}
}

type NetAddress struct {
	Host string
	Port int
}

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

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func (c *Config) parseFlags() {

	addr := new(NetAddress)
	_ = flag.Value(addr)

	urlAddr := new(NetAddress)
	_ = flag.Value(urlAddr)

	flag.Var(addr, "a", "Net address host:port")
	flag.Var(urlAddr, "b", "address and port for short URL")

	flag.Parse()

	// проверка значения addr, чтобы записать в переменную bindAddr
	if addr.String() != ":0" {
		c.bindAddr = addr.String()
	}

	// проверка значения urlAddr, чтобы записать в переменную shortURLAddr
	if urlAddr.String() != ":0" {
		c.ShortURLAddr = urlAddr.String()
	}
}
