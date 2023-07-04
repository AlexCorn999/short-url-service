package handlers

import (
	"fmt"
	"net/http"
)

// StringAccept принимает ссылку и возвращает закодированную ссылку
func StringAcceptAndBack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if r.Method == http.MethodPost {
		// парсинг тела запроса POST
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		body := ""
		for k := range r.Form {
			body += k
		}

		link := fmt.Sprintf("http:%s/%s", r.Host, body)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(link))
		return
	}

	if r.Method == http.MethodGet {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		body := r.URL.String()

		link := fmt.Sprintf("http:%s%s", r.Host, body)
		w.Header().Set("Location", link)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
}

/*
// StringBack возвращает ссылку по id
func StringBack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body := r.URL.String()

	link := fmt.Sprintf("http:%s%s", r.Host, body)
	w.Header().Set("Location", link)
	w.WriteHeader(http.StatusTemporaryRedirect)
}*/
