package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/go-chi/cors"
	"github.com/mrmod/terrastate/analyzer"
)

var (
	allowedOrigins = []string{"http://*"}
	allowedMethods = []string{"GET", "POST", "HEAD", "OPTIONS", "PUT", "DELETE"}
	allowedHeaders = []string{"Content-Type"}
)

type PlanState struct {
	analyzer.Plan
	cfg analyzer.PlanDatastoreConfig
}

// CreatePlan Creates a plan with a user-defined SessionId or 'default-session'
// storing the plan to a datastore and indexing the Session by its UTC create
// time as observed by the API
func (p *PlanState) CreatePlan(w http.ResponseWriter, req *http.Request) {
	// Many plans can come from the same planning session
	sessionId := chi.URLParam(req, "sessionId")
	if sessionId == "" {
		sessionId = "default-session"
	}
	var plan analyzer.Plan
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&plan); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	b, _ := json.Marshal(plan)
	changeSetSum := sha256.Sum256(b)

	plan.ChangeSetId = base64.URLEncoding.EncodeToString(changeSetSum[0:])
	plannedChange, err := analyzer.SavePlan(p.cfg, sessionId, plan)

	if err != nil {
		log.Printf("Plan failed to upload: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(plannedChange)
}

const (
	last7Days = iota
	last30Days
	last90Days
	last3Months = last90Days
	last6Months = iota
	nowThisMoment
)

var (
	timesInThePast = map[int]time.Time{
		nowThisMoment: time.Now().UTC(),
		last7Days:     time.Now().UTC().Add(-1 * 7 * 24 * time.Hour),
		last30Days:    time.Now().UTC().Add(-1 * 30 * 24 * time.Hour),
		last90Days:    time.Now().UTC().Add(-1 * 90 * 24 * time.Hour),
		last6Months:   time.Now().UTC().Add(-1 * 180 * 24 * time.Hour),
	}
)

// IndexPlans Provides a time-ordered list of Plan sessions
func (p *PlanState) IndexPlans(w http.ResponseWriter, req *http.Request) {
	// startOfWindow MUST be greater in magnitude and in the same direction as the endOfWindow time
	// Examples of valid times: (Now, Now-2Hours), (Now-30Days, Now-90Days),
	startOfWindow := timesInThePast[last90Days]
	endOfWindow := timesInThePast[last6Months]
	plans, err := analyzer.ListPlans(p.cfg, startOfWindow, endOfWindow)
	if err != nil {
		log.Printf("Failed to list plans: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(plans)
}

// ShowPlan Provides a plan for a specific 'changeSetId'
func (p *PlanState) ShowPlan(w http.ResponseWriter, req *http.Request) {
	changeSetId := chi.URLParam(req, "changeSetId")
	log.Printf("Plan changeSetId: %s", changeSetId)

	plan, err := analyzer.GetPlan(p.cfg, analyzer.ChangeSetId(changeSetId))
	if err != nil {
		log.Printf("Failed to get plan changeSet: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(plan)
}

type DiffApiResponse struct {
	Updates []analyzer.UpdatedResource `json:"updates"`
	Deletes []analyzer.ResourceChange  `json:"deletes"`
	Creates []analyzer.CreatedResource `json:"creates"`
}

func (p *PlanState) ShowDiff(w http.ResponseWriter, req *http.Request) {
	changeSetId := chi.URLParam(req, "changeSetId")
	log.Printf("Getting diff for ChangeSetId: %s", changeSetId)
	plan, err := analyzer.GetPlan(p.cfg, analyzer.ChangeSetId(changeSetId))
	if err != nil {
		log.Printf("Failed to get plan changeSet: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	apiResponse := DiffApiResponse{
		// TODO: These updated resources should include the Before and After data from the Terraform Plan
		Updates: analyzer.Diffs(analyzer.GetUpdatedResources(*plan)),
		// TODO: These deleted/destroyed resources should include the Before data from the Terraform Plan
		Deletes: analyzer.GetDeletedResources(*plan),
		Creates: analyzer.CreatedResources(*plan),
	}
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(apiResponse); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
func main() {

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   allowedMethods,
		AllowedHeaders:   allowedHeaders,
		AllowCredentials: true,
	}))
	configFile := "backend/config.dev.json"
	if _configFile, ok := os.LookupEnv("TERRASTATE_CONFIG"); ok {
		configFile = _configFile
	}

	planState := &PlanState{
		cfg: analyzer.PlanDatastoreConfig{
			BackendConfig: analyzer.LoadConfig(configFile),
			Config:        analyzer.NewAwsClient(),
		},
	}

	router.Route("/plans", func(_router chi.Router) {
		_router.Post("/", planState.CreatePlan)
		_router.Post("/{sessionId}", planState.CreatePlan)
		_router.Get("/{changeSetId}", planState.ShowPlan)
		_router.Get("/", planState.IndexPlans)
	})
	router.Route("/diffs", func(_router chi.Router) {
		_router.Get("/{changeSetId}", planState.ShowDiff)
	})
	port := "8000"
	if _port, ok := os.LookupEnv("PORT"); ok {
		if matched, err := regexp.MatchString("^[0-9]{2,4}$", _port); matched && err == nil {
			port = _port
		}
	}
	log.Printf("Starting server on port %s", port)
	http.ListenAndServe(":"+port, router)
}
