package core_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gustavooferreira/wcrawler/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebClient(t *testing.T) {
	tests := map[string]struct {
		path               string
		htmlBody           string
		expectedErr        bool
		expectedStatusCode int
		expectedLinks      []core.URLEntity
	}{
		"parse 1": {
			path:               "/random/path/to/oblivion/index.html",
			htmlBody:           htmlBody1,
			expectedStatusCode: 200,
			expectedLinks: []core.URLEntity{{
				Host: "www.example.com",
				Raw:  "http://www.example.com/file.html",
			}, {
				Host: "%s",
				Raw:  "%s/path/to/file999",
			}, {
				Host: "%s",
				Raw:  "%s/random/path/to/oblivion/path/to/file2",
			}},
			expectedErr: false,
		},
	}

	// Setup
	c := &http.Client{}
	wc := core.NewWebClient(c)

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, test.htmlBody)
			}))
			defer ts.Close()

			queryURL := ts.URL + test.path

			u, err := url.Parse(ts.URL)
			if err != nil {
				require.FailNow(t, "failed parsing fake server URL")
			}

			host := u.Host

			statusCode, links, err := wc.GetLinks(queryURL)

			if test.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Replace URLEntity's Host and Raw with the URL provided by test server
			for i, l := range test.expectedLinks {
				if strings.Contains(l.Host, "%s") {
					test.expectedLinks[i].Host = fmt.Sprintf(l.Host, host)
				}

				if strings.Contains(l.Raw, "%s") {
					test.expectedLinks[i].Raw = fmt.Sprintf(l.Raw, ts.URL)
				}
			}

			assert.Equal(t, test.expectedStatusCode, statusCode)
			assert.Equal(t, test.expectedLinks, links)
		})
	}
}
