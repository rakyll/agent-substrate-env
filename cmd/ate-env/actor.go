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
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/rakyll/agent-substrate-env/internal/session"
)

// newActorCmd builds the "actor" subcommand: the in-actor tool executor. It
// runs inside the sandboxed actor and executes incoming tool calls in-process
// against the local environment.
func newActorCmd() *cobra.Command {
	var configPath string
	cmd := &cobra.Command{
		Use:   "actor",
		Short: "Run the in-actor tool executor (execute tool calls)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runActor(configPath)
		},
	}
	cmd.Flags().StringVar(&configPath, "config", "config.yaml", "path to the YAML configuration file")
	return cmd
}

func runActor(configPath string) error {
	cfg, store, err := newSessionManager(configPath)
	if err != nil {
		return err
	}

	log.Printf("Starting Agent Substrate environment actor...")

	mux := http.NewServeMux()
	// Executing tool calls is the primary operation on a session, so it is a
	// POST to the session resource itself. Both the environment and the session
	// id live in the path, which the stateless executor needs to pick the tool
	// allowlist.
	mux.HandleFunc("POST /v1/environments/{env}/sessions/{session_id}", handleExecute(store))
	mux.HandleFunc("GET /healthz", handleHealthz)

	addr := listenAddr(cfg.Listen)
	log.Printf("Serving HTTP requests on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		return fmt.Errorf("HTTP server failed: %w", err)
	}
	return nil
}

// handleExecute handles session tool execution requests.
func handleExecute(store *session.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req session.ExecuteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid request payload: %v", err), http.StatusBadRequest)
			return
		}

		envName := r.PathValue("env")
		sessionID := r.PathValue("session_id")
		responses, err := store.Execute(r.Context(), sessionID, envName, req.EnvVariables, req.Inputs)
		if err != nil {
			log.Printf("failed to execute tool calls for session %s: %v", sessionID, err)
			http.Error(w, fmt.Sprintf("failed to execute tool calls: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(session.ExecuteResponse{Outputs: responses})
	}
}
