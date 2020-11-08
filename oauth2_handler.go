package oauth2

import (
	"context"
	"encoding/json"
	"github.com/common-go/auth"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

type OAuth2ActivityLogWriter interface {
	Write(ctx context.Context, resource string, action string, success bool, desc string) error
}

type OAuth2Handler struct {
	OAuth2Service OAuth2Service
	Resource      string
	Action        string
	LogError      func(context.Context, string)
	Ip            string
	LogWriter     OAuth2ActivityLogWriter
}

func NewOAuth2Handler(oAuth2IntegrationService OAuth2Service, resource string, action string, logError func(context.Context, string), ip string, logService OAuth2ActivityLogWriter) *OAuth2Handler {
	if len(resource) == 0 {
		resource = "oauth2"
	}
	if len(action) == 0 {
		action = "authenticate"
	}
	if len(ip) == 0 {
		ip = "ip"
	}
	return &OAuth2Handler{OAuth2Service: oAuth2IntegrationService, Resource: resource, Action: action, LogError: logError, Ip: ip, LogWriter: logService}
}
func (h *OAuth2Handler) Configuration(w http.ResponseWriter, r *http.Request) {
	id := ""
	if r.Method == "GET" {
		i := strings.LastIndex(r.RequestURI, "/")
		if i >= 0 {
			id = r.RequestURI[i+1:]
		}
	} else {
		b, er1 := ioutil.ReadAll(r.Body)
		if er1 != nil {
			respondString(w, r, http.StatusBadRequest, "Body cannot is empty")
			return
		}
		id = strings.Trim(string(b), " ")
	}
	if len(id) == 0 {
		respondString(w, r, http.StatusBadRequest, "request cannot is empty")
		return
	}
	model, err := h.OAuth2Service.Configuration(r.Context(), id)
	if err != nil {
		return
		if h.LogError != nil {
			h.LogError(r.Context(), err.Error())
		}
		respond(w, r, http.StatusOK, nil, h.LogWriter, h.Resource, h.Action, false, err.Error())
	} else {
		respond(w, r, http.StatusOK, model, h.LogWriter, h.Resource, "configuration", true, "")
	}
}
func (h *OAuth2Handler) Configurations(w http.ResponseWriter, r *http.Request) {
	model, err := h.OAuth2Service.Configurations(r.Context())
	if err != nil {
		return
		if h.LogError != nil {
			h.LogError(r.Context(), err.Error())
		}
		respond(w, r, http.StatusOK, nil, h.LogWriter, h.Resource, h.Action, false, err.Error())
	} else {
		respond(w, r, http.StatusOK, model, h.LogWriter, h.Resource, "configuration", true, "")
	}
}
func (h *OAuth2Handler) Authenticate(w http.ResponseWriter, r *http.Request) {
	var request OAuth2Info
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		respondString(w, r, http.StatusBadRequest, "Cannot decode OAuth2Info model: "+err.Error())
		return
	}
	var authorization string
	if len(r.Header["Authorization"]) < 1 {
		authorization = ""
	} else {
		authorization = r.Header["Authorization"][0]
	}
	ip := getRemoteIp(r)
	var ctx context.Context
	ctx = r.Context()
	if len(h.Ip) > 0 {
		ctx = context.WithValue(ctx, h.Ip, ip)
		r = r.WithContext(ctx)
	}
	result, err := h.OAuth2Service.Authenticate(r.Context(), request, authorization)
	if err != nil {
		result.Status = auth.StatusSystemError
		if h.LogError != nil {
			h.LogError(r.Context(), err.Error())
		}
		respond(w, r, http.StatusOK, result, h.LogWriter, h.Resource, h.Action, false, err.Error())
	} else {
		respond(w, r, http.StatusOK, result, h.LogWriter, h.Resource, h.Action, true, "")
	}
}

func respond(w http.ResponseWriter, r *http.Request, code int, result interface{}, logWriter OAuth2ActivityLogWriter, resource string, action string, success bool, desc string) {
	response, _ := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
	if logWriter != nil {
		newCtx := context.WithValue(r.Context(), "request", r)
		logWriter.Write(newCtx, resource, action, success, desc)
	}
}
func respondString(w http.ResponseWriter, r *http.Request, code int, result string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(result))
}
func getRemoteIp(r *http.Request) string {
	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteIP = r.RemoteAddr
	}
	return remoteIP
}
