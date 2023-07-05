package apiserver

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AlexCorn999/short-url-service/internal/app/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringAccept(t *testing.T) {
	server := APIServer{}
	server.configureRouter()
	server.storage = *store.NewStorage()

	type want struct {
		statusCode int
		response   string
	}

	testTable := []struct {
		request string
		body    string
		want    want
	}{
		{
			request: "/",
			body:    "Yandex.ru",
			want: want{
				statusCode: 201,
				response:   "http://example.com/1",
			},
		},
		{
			request: "/22",
			body:    "Yandex.ru",
			want: want{
				statusCode: 400,
				response:   "",
			},
		},
		{
			request: "/",
			body:    "http://Skillbox.ru",
			want: want{
				statusCode: 201,
				response:   "http://example.com/2",
			},
		},
	}

	for _, tc := range testTable {
		req := httptest.NewRequest(http.MethodPost, tc.request, strings.NewReader(tc.body))
		w := httptest.NewRecorder()
		server.StringAcceptAndBack(w, req)

		result := w.Result()
		defer result.Body.Close()
		body, err := io.ReadAll(result.Body)
		require.NoError(t, err)

		assert.Equal(t, tc.want.statusCode, result.StatusCode)
		assert.Equal(t, tc.want.response, string(body))
	}
}

func TestStringBack(t *testing.T) {
	server := APIServer{}
	server.configureRouter()
	server.storage = *store.NewStorage()

	server.storage.Data["1"] = "Yandex.ru"
	server.storage.Data["2"] = "http://Skillbox.ru"

	type want struct {
		statusCode  int
		contentType string
	}

	testTable := []struct {
		request string
		want    want
	}{
		{
			request: "/1",
			want: want{
				statusCode:  307,
				contentType: "Yandex.ru",
			},
		},
		{
			request: "/2",
			want: want{
				statusCode:  307,
				contentType: "http://Skillbox.ru",
			},
		},
		{
			request: "/3",
			want: want{
				statusCode:  400,
				contentType: "",
			},
		},
	}

	for _, tc := range testTable {
		req := httptest.NewRequest(http.MethodGet, tc.request, nil)
		w := httptest.NewRecorder()
		server.StringAcceptAndBack(w, req)

		result := w.Result()

		assert.Equal(t, tc.want.statusCode, result.StatusCode)
		assert.Equal(t, tc.want.contentType, result.Header.Get("Location"))
	}
}
