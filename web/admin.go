package web

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/daohoangson/go-socialcounters/utils"
)

func ConfigGet(u utils.Utils, w http.ResponseWriter, r *http.Request) {
	specifiedSecret := parseSecret(r)
	configSecret := os.Getenv("CONFIG_SECRET")
	if configSecret == "" || specifiedSecret != configSecret {
		w.WriteHeader(http.StatusForbidden)
		writeJson(u, w, r, "{}")
		u.Errorf("admin.ConfigGet: wrong secret %s", specifiedSecret)
		return
	}

	keys := parseKeys(r)
	values := make(map[string]string)
	for _, key := range keys {
		values[key] = u.ConfigGet(key)
	}

	valuesJson, err := json.Marshal(values)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJson(u, w, r, "{}")
		u.Errorf("admin.ConfigGet: json.Marshal(values) error %v", err)
		return
	}

	writeJson(u, w, r, string(valuesJson))
}

func ConfigPost(u utils.Utils, w http.ResponseWriter, r *http.Request) {
	specifiedSecret := parseSecret(r)
	configSecret := os.Getenv("CONFIG_SECRET")
	if configSecret == "" || specifiedSecret != configSecret {
		w.WriteHeader(http.StatusForbidden)
		u.Errorf("admin.ConfigGet: wrong secret %s", specifiedSecret)
		return
	}

	r.ParseForm()
	if keys, ok := r.PostForm["key"]; ok {
		for _, key := range keys {
			if len(key) < 1 {
				continue
			}

			if values, ok := r.PostForm[key]; ok {
				for _, value := range values {
					if err := u.ConfigSet(key, value); err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						u.Errorf("admin.ConfigSet: u.ConfigSet(%q, %q) error %v", key, value, err)
						return
					}
				}
			}
		}
	}

	RulesRefresh(u)
	w.WriteHeader(http.StatusAccepted)
}

func parseSecret(r *http.Request) string {
	q := r.URL.Query()
	if secrets, ok := q["secret"]; ok {
		return secrets[0]
	}

	return ""
}

func parseKeys(r *http.Request) []string {
	q := r.URL.Query()
	if keys, ok := q["key"]; ok {
		return keys
	}

	return []string{}
}
