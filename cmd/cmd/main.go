package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/arham-abiyan/reconciliation/internal/services/reconciliation"
)

// Define a custom type for the array of strings
type stringArray []string

// Implement the `String` method for the `flag.Value` interface
func (s *stringArray) String() string {
	return strings.Join(*s, ",")
}

// Implement the `Set` method for the `flag.Value` interface
func (s *stringArray) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	// Define a string array flag
	var system, bank, startDate, endDate stringArray
	flag.Var(&system, "system", "Specify file path for system transactions")
	flag.Var(&bank, "bank", "Specify file paths (can be used multiple times) for bank transactions")
	flag.Var(&startDate, "start", "Specify start date")
	flag.Var(&endDate, "end", "Specify end date")

	// Parse the command-line flags
	flag.Parse()

	svc := reconciliation.New(bank, system[0], startDate[0], endDate[0])

	result, err := svc.Reconcile()
	if err != nil {
		log.Fatal(err)
	}

	// Print reconciliation summary
	fmt.Println("Reconciliation Summary")
	fmt.Println("-----------------------")
	fmt.Printf("Total transactions processed: %d\n", result.TotalProcessed)
	fmt.Printf("Total matched transactions: %d\n", result.Matched)
	fmt.Printf("Total unmatched transactions: %d\n", result.Unmatched)
	fmt.Printf("Total discrepancies: %.2f\n", result.Discrepancies)
	fmt.Println("\nUnmatched System Transactions:")
	for _, tx := range result.UnmatchedSystem {
		fmt.Println(tx)
	}
	fmt.Println("\nUnmatched By Bank Transactions:")
	for key, transactions := range result.UnmatchedByBank {
		fmt.Println(key)
		for _, tx := range transactions {
			fmt.Println(tx)
		}
	}
}
