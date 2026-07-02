// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Parse configurations from flags or environment variables
	port := flag.String("port", getEnv("PORT", "8080"), "Port for the HTTP environment service to listen on")
	ateapiAddr := flag.String("ateapi", getEnv("ATEAPI_ADDR", "localhost:8080"), "Address of the Agent Substrate ateapi gRPC server")
	atenetAddr := flag.String("atenet", getEnv("ATENET_ADDR", "localhost:8000"), "Address of the Agent Substrate atenet HTTP router")
	ateNamespace := flag.String("namespace", getEnv("ATE_NAMESPACE", "default"), "Agent Substrate namespace to create/resume actors")

	flag.Parse()

	log.Printf("Starting Agent Substrate environment service...")
	log.Printf("Listening Port: %s", *port)

	store := NewSessionStore(*ateapiAddr, *atenetAddr, *ateNamespace)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /environment/resume", handleResume(store))
	mux.HandleFunc("POST /environment/suspend", handleSuspend(store))
	mux.HandleFunc("POST /environment", handleExecute(store))
	mux.HandleFunc("GET /healthz", handleHealthz)

	log.Printf("Serving HTTP requests on :%s", *port)
	if err := http.ListenAndServe(":"+*port, mux); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}

// handleResume handles environment resume requests.
func handleResume(store *SessionStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ResumeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid request payload: %v", err), http.StatusBadRequest)
			return
		}

		if err := store.Resume(r.Context(), req); err != nil {
			log.Printf("failed to resume session %s: %v", req.SessionID, err)
			http.Error(w, fmt.Sprintf("failed to resume session: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

// handleSuspend handles environment suspend requests.
func handleSuspend(store *SessionStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SuspendRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid request payload: %v", err), http.StatusBadRequest)
			return
		}

		if err := store.Suspend(r.Context(), req.SessionID); err != nil {
			log.Printf("failed to suspend session %s: %v", req.SessionID, err)
			http.Error(w, fmt.Sprintf("failed to suspend session: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

// handleExecute handles environment tool execution requests.
func handleExecute(store *SessionStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ExecuteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid request payload: %v", err), http.StatusBadRequest)
			return
		}

		responses, err := store.Execute(r.Context(), req.SessionID, req.Inputs)
		if err != nil {
			log.Printf("failed to execute tool calls for session %s: %v", req.SessionID, err)
			http.Error(w, fmt.Sprintf("failed to execute tool calls: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(responses)
	}
}

// handleHealthz handles health check requests.
func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// getEnv gets an environment variable, returning fallback if empty.
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
