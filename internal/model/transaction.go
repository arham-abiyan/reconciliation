package model

import "time"

// Transaction represents a system transaction record
// TrxID: Unique identifier for the transaction
// Amount: Transaction amount
// Type: Type of transaction (DEBIT or CREDIT)
// TransactionTime: Date and time of the transaction
type Transaction struct {
	TransactionTime time.Time `json:"transaction_time"`
	TrxID           string    `json:"trx_id"`
	Type            string    `json:"type"`
	Amount          float64   `json:"amount"`
}

// BankStatement represents a bank statement record
// UniqueIdentifier: Unique identifier from the bank statement (can vary by bank)
// Amount: Transaction amount (negative for debits)
// Date: Date of the transaction
// Type: Type of transaction (DEBIT or CREDIT)
// Bank: Unmatched BankTransaction
type BankStatement struct {
	Date             time.Time `json:"date"`
	UniqueIdentifier string    `json:"unique_identifier"`
	Type             string    `json:"type"`
	Amount           float64   `json:"amount"`
	Bank             string    `json:"bank"`
}

type ReconcileResponse struct {
	UnmatchedSystem []Transaction              `json:"umatched_system"`
	UnmatchedByBank map[string][]BankStatement `json:"unmatched_by_bank"`
	Discrepancies   float64                    `json:"discrepancies"`
	TotalProcessed  int                        `json:"total_processed"`
	Matched         int                        `json:"matched"`
	Unmatched       int                        `json:"umatched"`
}
