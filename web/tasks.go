package web

import (
	"encoding/json"
	"net/http"

	"github.com/daohoangson/go-socialcounters/services"
	"github.com/daohoangson/go-socialcounters/utils"
)

func TaskRefresh(u utils.Utils, w http.ResponseWriter, r *http.Request) {
	dataMap := make(services.MapUrlServiceCount)
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&dataMap); err != nil {
		u.Errorf("web.TaskRefresh: decoder.Decode error %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	services.Refresh(u, &dataMap, parseTtl(u, r))
}