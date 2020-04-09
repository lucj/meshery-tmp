package handlers

import (
	"fmt"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/gorilla/sessions"
	"github.com/layer5io/meshery/models"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// FetchResultsHandler fetchs pages of results from SaaS and presents it to the UI
func (h *Handler) FetchResultsHandler(w http.ResponseWriter, req *http.Request, session *sessions.Session, _ *models.Preference, user *models.User, p models.Provider) {
	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// TODO: may be force login if token not found?????

	err := req.ParseForm()
	if err != nil {
		logrus.Errorf("Error: unable to parse form: %v", err)
		http.Error(w, "unable to process the received data", http.StatusForbidden)
		return
	}
	q := req.Form

	bdr, err := p.FetchResults(req, q.Get("page"), q.Get("pageSize"), q.Get("search"), q.Get("order"))
	if err != nil {
		http.Error(w, "error while getting load test results", http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	_, _ = w.Write(bdr)
}

// GetResultHandler gets an individual result from provider
func (h *Handler) GetResultHandler(w http.ResponseWriter, req *http.Request, session *sessions.Session, _ *models.Preference, user *models.User, p models.Provider) {
	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// TODO: may be force login if token not found?????
	id := req.URL.Query().Get("id")
	if id == "" {
		logrus.Errorf("Error: no id provided to get result")
		http.Error(w, "please provide a result id", http.StatusBadRequest)
		return
	}
	key := uuid.FromStringOrNil(id)
	if key == uuid.Nil {
		logrus.Errorf("Error: invalid id provided to get result")
		http.Error(w, "please provide a valid result id", http.StatusBadRequest)
		return
	}

	bdr, err := p.GetResult(req, key)
	if err != nil {
		http.Error(w, "error while getting load test results", http.StatusInternalServerError)
		return
	}
	sp, err := bdr.ConvertToSpec()
	if err != nil {
		http.Error(w, "error while getting load test results", http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/yaml")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="result_%s.yaml"`, bdr.ID))
	b, err := yaml.Marshal(sp)
	if err != nil {
		logrus.Errorf("Error: unable to marshal result: %v", err)
		http.Error(w, "error while getting test result", http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(b)
}
