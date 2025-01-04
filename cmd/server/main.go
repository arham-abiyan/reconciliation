package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/arham09/reconciliation-svc/internal/model"
	"github.com/arham09/reconciliation-svc/internal/services/reconciliation"
	"github.com/arham09/reconciliation-svc/pkg"
)

const (
	maxUploadSize = 10 << 20 // 10 MB
	uploadsDir    = "./uploads"
	port          = ":8080"
)

type APIResponse struct {
	Success bool                     `json:"success"`
	Data    *model.ReconcileResponse `json:"data"`
	Error   string                   `json:"error,omitempty"`
}

func main() {
	if err := os.MkdirAll(uploadsDir, os.ModePerm); err != nil {
		log.Fatal("Failed to create uploads directory:", err)
	}

	http.HandleFunc("/api/reconcile", handleReconciliation)

	log.Println("Server starting on...", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func sendJSONResponse(w http.ResponseWriter, statusCode int, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func handleReconciliation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONResponse(w, http.StatusMethodNotAllowed, APIResponse{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	// Validate content type
	if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
		sendJSONResponse(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Content-Type must be multipart/form-data",
		})
		return
	}

	// Limit request size
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Request too large. Max size is 10MB",
		})
		return
	}
	defer r.MultipartForm.RemoveAll()

	// Validate and parse dates
	startDate := r.FormValue("start_date")
	endDate := r.FormValue("end_date")
	if err := pkg.ValidateDates(startDate, endDate); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Handle system transaction file
	systemFile, systemHeader, err := r.FormFile("system_file")
	if err != nil {
		sendJSONResponse(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "System transaction file is required",
		})
		return
	}
	defer systemFile.Close()

	if err := pkg.ValidateFile(systemHeader); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Failed to process system file",
		})
		return
	}

	systemTransaction, err := pkg.SaveFile(systemHeader, uploadsDir, "system")
	if err != nil {
		sendJSONResponse(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Failed to process system file",
		})
		return
	}

	// Handle bank transaction files
	bankFiles := r.MultipartForm.File["bank_files"]
	if len(bankFiles) == 0 {
		sendJSONResponse(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "At least one bank transaction file is required",
		})
		return
	}

	bankTransactions := make([]string, 0, len(bankFiles))
	for _, fileHeader := range bankFiles {
		if err := pkg.ValidateFile(fileHeader); err != nil {
			sendJSONResponse(w, http.StatusBadRequest, APIResponse{
				Success: false,
				Error:   fmt.Sprintf("Error processing bank file %s: %v", fileHeader.Filename, err),
			})
			return
		}

		bankTransaction, err := pkg.SaveFile(fileHeader, uploadsDir, "bank")
		if err != nil {
			sendJSONResponse(w, http.StatusBadRequest, APIResponse{
				Success: false,
				Error:   fmt.Sprintf("Error processing bank file %s: %v", fileHeader.Filename, err),
			})
			return
		}
		bankTransactions = append(bankTransactions, bankTransaction)
	}

	svc := reconciliation.New(bankTransactions, systemTransaction, startDate, endDate)
	result, err := svc.Reconcile()
	if err != nil {
		sendJSONResponse(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Error reconcile transaction: %v", err),
		})
		return
	}

	sendJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    &result,
	})
}
