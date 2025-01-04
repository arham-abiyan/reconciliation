package reconciliation

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/arham09/reconciliation-svc/internal/model"
	"github.com/arham09/reconciliation-svc/internal/services"
)

type Service struct {
	bankCSV                       []string
	systemCSV, startDate, endDate string
}

var _ services.Reconciliation = (*Service)(nil)

func New(bankCSV []string, systemCSV, startDate, endDate string) *Service {
	return &Service{
		bankCSV:   bankCSV,
		systemCSV: systemCSV,
		startDate: startDate,
		endDate:   endDate,
	}
}

func (s *Service) Reconcile() (model.ReconcileResponse, error) {
	systemTransactions, _, err := parseCSV(s.systemCSV, true)
	if err != nil {
		fmt.Println("Error parsing system transactions:", err)
		return model.ReconcileResponse{}, err
	}

	var allBankStatements []model.BankStatement
	for _, bankCSV := range s.bankCSV {
		_, bankStatements, err := parseCSV(bankCSV, false)
		if err != nil {
			fmt.Println("Error parsing bank statement:", err)
			return model.ReconcileResponse{}, err
		}
		allBankStatements = append(allBankStatements, bankStatements...)
	}

	// Filter transactions within the specified date range
	filteredSystemTransactions := filterTransactions(systemTransactions, s.startDate, s.endDate, func(tx model.Transaction) time.Time {
		return tx.TransactionTime
	})
	filteredBankStatements := filterTransactions(allBankStatements, s.startDate, s.endDate, func(tx model.BankStatement) time.Time {
		return tx.Date
	})

	// Perform reconciliation
	return reconcileTransactions(filteredSystemTransactions, filteredBankStatements), nil
}

// parseCSV parses a CSV file into either system transactions or bank statements based on the isSystem flag
func parseCSV(filePath string, isSystem bool) ([]model.Transaction, []model.BankStatement, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	fileName := extractBaseName(filePath)

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}

	if isSystem {
		transactions := make([]model.Transaction, 0, len(records)-1)
		for _, record := range records[1:] {
			amount, _ := strconv.ParseFloat(record[1], 64)
			trxTime, _ := time.Parse("2006-01-02 15:04:05", record[3])
			transactions = append(transactions, model.Transaction{
				TrxID:           record[0],
				Amount:          amount,
				Type:            record[2],
				TransactionTime: trxTime,
			})
		}

		return transactions, nil, nil
	}

	bankStatements := make([]model.BankStatement, 0, len(records)-1)
	for _, record := range records[1:] {
		amount, _ := strconv.ParseFloat(record[1], 64)
		date, _ := time.Parse("2006-01-02", record[2])

		trxType := "DEBIT"
		if amount < 0 {
			trxType = "CREDIT"
		}

		bankStatements = append(bankStatements, model.BankStatement{
			UniqueIdentifier: record[0],
			Amount:           math.Abs(amount),
			Type:             trxType,
			Date:             date,
			Bank:             fileName,
		})
	}

	return nil, bankStatements, nil
}

// filterTransactions filters transactions within a specified date range
// extractDate: A function that extracts the date from the transaction struct
// used generic to make system transaction and bank transaction as allowed input
func filterTransactions[T any](transactions []T, startDateStr, endDateStr string, extractDate func(T) time.Time) []T {
	var filtered []T

	startDate, _ := time.Parse("2006-01-02", startDateStr)
	endDate, _ := time.Parse("2006-01-02", endDateStr)
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, time.UTC)

	for _, tx := range transactions {
		txDate := extractDate(tx)
		if txDate.After(startDate) && txDate.Before(endDate) {
			filtered = append(filtered, tx)
		}
	}

	return filtered
}

// absDiff calculates the absolute difference between two float64 values
func absDiff(a, b float64) float64 {
	if a > b {
		return a - b
	}
	return b - a
}

// reconcileTransactions matches system transactions with bank statements
// Returns counts of processed, matched, and unmatched transactions, along with discrepancies and unmatched records
func reconcileTransactions(systemTransactions []model.Transaction, bankStatements []model.BankStatement) model.ReconcileResponse {
	matched := 0
	discrepancies := 0.0
	totalProcessed := 0
	unmatchedSystem := make([]model.Transaction, 0, len(systemTransactions))
	unmatchedByBank := make(map[string][]model.BankStatement)
	bankMap := make(map[string]model.BankStatement)

	// Create a map of bank transactions for O(1) lookup
	for _, bankTx := range bankStatements {
		key := bankTx.UniqueIdentifier
		bankMap[key] = bankTx
	}

	// Match system transactions with bank transactions
	for _, sysTx := range systemTransactions {
		key := sysTx.TrxID
		totalProcessed++
		if bankEntries, exists := bankMap[key]; exists {
			matched++
			discrepancies += absDiff(sysTx.Amount, bankEntries.Amount)
			delete(bankMap, key)
		} else {
			unmatchedSystem = append(unmatchedSystem, sysTx)
		}
	}

	// Collect unmatched bank transactions
	for _, bankEntries := range bankMap {
		_, ok := unmatchedByBank[bankEntries.Bank]
		if ok {
			unmatchedByBank[bankEntries.Bank] = append(unmatchedByBank[bankEntries.Bank], bankEntries)
		}

		unmatchedByBank[bankEntries.Bank] = []model.BankStatement{bankEntries}
	}

	return model.ReconcileResponse{
		UnmatchedSystem: unmatchedSystem,
		Discrepancies:   discrepancies,
		TotalProcessed:  totalProcessed,
		Matched:         matched,
		UnmatchedByBank: unmatchedByBank,
		Unmatched:       len(bankMap) + len(unmatchedSystem),
	}
}

// extractBaseName extracts the base name from a file path or file name.
// It removes the directory path, trims the ".csv" extension, and returns the last
func extractBaseName(filename string) string {
	base := filepath.Base(filename)
	withoutExt := strings.TrimSuffix(base, ".csv")
	parts := strings.Split(withoutExt, "_")

	if len(parts) > 1 {
		return parts[len(parts)-1]
	}

	return withoutExt
}
