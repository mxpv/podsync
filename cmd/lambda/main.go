package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var (
	sess = session.Must(session.NewSession())
	svc  = lambda.New(sess)
)

type response struct {
	StatusCode int               `json:"statusCode"`
	StatusDesc string            `json:"statusDescription"`
	Headers    map[string]string `json:"headers"`
}

func resolve(w http.ResponseWriter, r *http.Request) {
	var (
		feedID  = chi.URLParam(r, "feedID")
		videoID = chi.URLParam(r, "videoID")
		path    = fmt.Sprintf("/download/%s/%s", feedID, videoID)
		params  = map[string]string{"path": path}
	)

	// Serialize lambda request payload
	payload, err := json.Marshal(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	request := &lambda.InvokeInput{}
	request.SetPayload(payload)
	request.SetFunctionName("Resolver")
	request.SetQualifier("PROD")

	// Invoke lambda function
	output, err := svc.Invoke(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Deserialize lambda response
	var out response
	if err := json.Unmarshal(output.Payload, &out); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write response
	for name, value := range out.Headers {
		w.Header().Set(name, value)
	}
	w.WriteHeader(out.StatusCode)
	fmt.Fprintln(w, out.StatusDesc)
}

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.GetHead)

	r.Use(middleware.Timeout(15 * time.Second))

	r.Route("/download", func(r chi.Router) {
		r.Get("/{feedID}/{videoID}", resolve)
	})

	http.ListenAndServe(":5002", r)
}
