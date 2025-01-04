package reconciliation

import (
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/arham09/reconciliation-svc/internal/model"
)

func parseDateWithTime(dateStr string) time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", dateStr)
	return t
}

func parseDate(dateStr string) time.Time {
	parsed, _ := time.Parse("2006-01-02", dateStr)
	return parsed
}

func compareTransactions(a, b model.Transaction) bool {
	return a.TrxID == b.TrxID &&
		a.Amount == b.Amount &&
		a.Type == b.Type &&
		a.TransactionTime.Equal(b.TransactionTime)
}

func compareBankStatements(a, b model.BankStatement) bool {
	return a.UniqueIdentifier == b.UniqueIdentifier &&
		a.Amount == b.Amount &&
		a.Type == b.Type &&
		a.Date.Equal(b.Date)
}

func TestFilterTransactions(t *testing.T) {
	systemTransactions := []model.Transaction{
		{TrxID: "T1", Amount: 100.0, Type: "DEBIT", TransactionTime: parseDateWithTime("2024-12-02 08:00:00")},
		{TrxID: "T2", Amount: 200.0, Type: "CREDIT", TransactionTime: parseDateWithTime("2024-12-02 18:00:00")},
		{TrxID: "T3", Amount: 300.0, Type: "DEBIT", TransactionTime: parseDateWithTime("2024-12-01 08:00:00")},
		{TrxID: "T4", Amount: 400.0, Type: "CREDIT", TransactionTime: parseDateWithTime("2024-12-02 08:00:00")},
	}

	testCases := []struct {
		name, startDate, endDate string
		expected                 int
	}{
		{
			name:      "Filter transactions within 2024-01-02 to 2024-01-03",
			startDate: "2024-01-02",
			endDate:   "2024-01-03",
			expected:  0,
		},
		{
			name:      "Filter transactions on a single day 2024-01-01",
			startDate: "2024-01-01",
			endDate:   "2024-0591-01",
			expected:  0,
		},
		{
			name:      "Filter transactions for a range 2024-12-01 to 2024-12-02",
			startDate: "2024-12-01",
			endDate:   "2024-12-02",
			expected:  4,
		},
		{
			name:      "Filter transactions for 2024-12-02",
			startDate: "2024-12-02",
			endDate:   "2024-12-02",
			expected:  3,
		},
		{
			name:      "Filter transactions for 2024-12-01",
			startDate: "2024-12-01",
			endDate:   "2024-12-01",
			expected:  1,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filtered := filterTransactions(systemTransactions, tc.startDate, tc.endDate, func(tx model.Transaction) time.Time {
				return tx.TransactionTime
			})

			if len(filtered) != tc.expected {
				t.Errorf("Expected %d filtered transactions, got %d", tc.expected, len(filtered))
			}
		})
	}
}

func TestReconcileTransactions(t *testing.T) {
	tests := []struct {
		name              string
		systemTrx         []model.Transaction
		bankStmt          []model.BankStatement
		wantTotal         int
		wantMatched       int
		wantUnmatched     int
		wantUnmatchedSys  int
		wantUnmatchedBank int
		wantDiscrepancies float64
	}{
		{
			name: "discrepancy",
			systemTrx: []model.Transaction{
				{TrxID: "T1", Amount: 100.0, Type: "DEBIT", TransactionTime: parseDate("2024-01-01").Add(10 * time.Hour)},
				{TrxID: "T2", Amount: 200.0, Type: "CREDIT", TransactionTime: parseDate("2024-01-02").Add(14 * time.Hour)},
				{TrxID: "T3", Amount: 300.0, Type: "DEBIT", TransactionTime: parseDate("2024-01-03").Add(16 * time.Hour)},
			},
			bankStmt: []model.BankStatement{
				{UniqueIdentifier: "T1", Amount: 100.0, Date: parseDate("2024-01-01")},
				{UniqueIdentifier: "T2", Amount: 250.0, Date: parseDate("2024-01-02")},
				{UniqueIdentifier: "B3", Amount: 300.0, Date: parseDate("2024-01-03")},
			},
			wantTotal:         3,
			wantMatched:       2,
			wantUnmatched:     2,
			wantUnmatchedSys:  1,
			wantUnmatchedBank: 1,
			wantDiscrepancies: 50.0,
		},
		{
			name: "empty",
			systemTrx: []model.Transaction{
				{TrxID: "T1", Amount: 100.0, Type: "DEBIT", TransactionTime: parseDate("2024-01-01")},
				{TrxID: "T2", Amount: 200.0, Type: "CREDIT", TransactionTime: parseDate("2024-01-02")},
			},
			bankStmt:          []model.BankStatement{},
			wantTotal:         2,
			wantMatched:       0,
			wantUnmatched:     2,
			wantUnmatchedSys:  2,
			wantDiscrepancies: 0.0,
		},
		{
			name: "ok",
			systemTrx: []model.Transaction{
				{TrxID: "T1", Amount: 100.0, Type: "DEBIT", TransactionTime: parseDate("2024-01-01")},
				{TrxID: "T2", Amount: 200.0, Type: "CREDIT", TransactionTime: parseDate("2024-01-02")},
			},
			bankStmt: []model.BankStatement{
				{UniqueIdentifier: "T1", Amount: 100.0, Date: parseDate("2024-01-01")},
				{UniqueIdentifier: "T2", Amount: 200.0, Date: parseDate("2024-01-02")},
			},
			wantTotal:         2,
			wantMatched:       2,
			wantUnmatched:     0,
			wantUnmatchedSys:  0,
			wantUnmatchedBank: 0,
			wantDiscrepancies: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reconcileTransactions(tt.systemTrx, tt.bankStmt)

			if result.TotalProcessed != tt.wantTotal {
				t.Errorf("totalProcessed = %d, want %d", result.TotalProcessed, tt.wantTotal)
			}
			if result.Matched != tt.wantMatched {
				t.Errorf("matched = %d, want %d", result.Matched, tt.wantMatched)
			}
			if result.Unmatched != tt.wantUnmatched {
				t.Errorf("unmatched = %d, want %d", result.Unmatched, tt.wantUnmatched)
			}
			if len(result.UnmatchedSystem) != tt.wantUnmatchedSys {
				t.Errorf("unmatched system transactions = %d, want %d",
					len(result.UnmatchedSystem), tt.wantUnmatchedSys)
			}
			if len(result.UnmatchedByBank) != tt.wantUnmatchedBank {
				t.Errorf("unmatched by bank transactions = %d, want %d",
					len(result.UnmatchedByBank), tt.wantUnmatchedBank)
			}
			if math.Abs(result.Discrepancies-tt.wantDiscrepancies) > 0.01 {
				t.Errorf("discrepancies = %.2f, want %.2f",
					result.Discrepancies, tt.wantDiscrepancies)
			}
		})
	}
}

func TestAbsDiff(t *testing.T) {
	if absDiff(100.0, 80.0) != 20.0 {
		t.Errorf("Expected absDiff(100.0, 80.0) = 20.0")
	}
	if absDiff(80.0, 100.0) != 20.0 {
		t.Errorf("Expected absDiff(80.0, 100.0) = 20.0")
	}
	if absDiff(100.0, 100.0) != 0.0 {
		t.Errorf("Expected absDiff(100.0, 100.0) = 0.0")
	}
}

func TestParseCSV(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()

	// Test case 1: System transactions CSV
	systemCSVContent := `TrxID,Amount,Type,TransactionTime
T1,100.00,DEBIT,2024-01-01 10:30:00
T2,200.00,CREDIT,2024-01-02 14:45:00
T3,300.00,DEBIT,2024-01-03 16:15:00`

	systemFilePath := filepath.Join(tmpDir, "system.csv")
	if err := os.WriteFile(systemFilePath, []byte(systemCSVContent), 0644); err != nil {
		t.Fatalf("Failed to create system test file: %v", err)
	}

	// Test case 2: Bank statements CSV
	bankCSVContent := `UniqueIdentifier,Amount,Date
T1,100.00,2024-01-01
T2,-250.00,2024-01-02
T3,300.00,2024-01-03`

	bankFilePath := filepath.Join(tmpDir, "bank.csv")
	if err := os.WriteFile(bankFilePath, []byte(bankCSVContent), 0644); err != nil {
		t.Fatalf("Failed to create bank test file: %v", err)
	}

	// Test case 3: Invalid file path
	invalidPath := filepath.Join(tmpDir, "nonexistent.csv")

	tests := []struct {
		name         string
		filePath     string
		isSystem     bool
		wantErr      bool
		wantSysTrx   []model.Transaction
		wantBankStmt []model.BankStatement
	}{
		{
			name:     "Valid system transactions",
			filePath: systemFilePath,
			isSystem: true,
			wantErr:  false,
			wantSysTrx: []model.Transaction{
				{
					TrxID:           "T1",
					Amount:          100.00,
					Type:            "DEBIT",
					TransactionTime: parseDateWithTime("2024-01-01 10:30:00"),
				},
				{
					TrxID:           "T2",
					Amount:          200.00,
					Type:            "CREDIT",
					TransactionTime: parseDateWithTime("2024-01-02 14:45:00"),
				},
				{
					TrxID:           "T3",
					Amount:          300.00,
					Type:            "DEBIT",
					TransactionTime: parseDateWithTime("2024-01-03 16:15:00"),
				},
			},
		},
		{
			name:     "Valid bank statements",
			filePath: bankFilePath,
			isSystem: false,
			wantErr:  false,
			wantBankStmt: []model.BankStatement{
				{
					UniqueIdentifier: "T1",
					Amount:           100.00,
					Type:             "DEBIT",
					Date:             parseDate("2024-01-01"),
				},
				{
					UniqueIdentifier: "T2",
					Amount:           250.00,
					Type:             "CREDIT",
					Date:             parseDate("2024-01-02"),
				},
				{
					UniqueIdentifier: "T3",
					Amount:           300.00,
					Type:             "DEBIT",
					Date:             parseDate("2024-01-03"),
				},
			},
		},
		{
			name:     "Invalid file path",
			filePath: invalidPath,
			isSystem: true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sysTrx, bankStmt, err := parseCSV(tt.filePath, tt.isSystem)

			// Check error condition
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCSV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Check system transactions
			if tt.isSystem {
				if len(sysTrx) != len(tt.wantSysTrx) {
					t.Errorf("parseCSV() got %d transactions, want %d", len(sysTrx), len(tt.wantSysTrx))
					return
				}

				for i, got := range sysTrx {
					want := tt.wantSysTrx[i]
					if !compareTransactions(got, want) {
						t.Errorf("Transaction %d mismatch:\ngot: %+v\nwant: %+v", i, got, want)
					}
				}
			} else {
				// Check bank statements
				if len(bankStmt) != len(tt.wantBankStmt) {
					t.Errorf("parseCSV() got %d statements, want %d", len(bankStmt), len(tt.wantBankStmt))
					return
				}

				for i, got := range bankStmt {
					want := tt.wantBankStmt[i]
					if !compareBankStatements(got, want) {
						t.Errorf("Bank statement %d mismatch:\ngot: %+v\nwant: %+v", i, got, want)
					}
				}
			}
		})
	}
}
