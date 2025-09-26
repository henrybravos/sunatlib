package main

import (
	"fmt"
	"log"
	"os"

	"github.com/henrybravos/sunatlib"
)

func validateExample() {
	// Example 1: Using the default validation client with master credentials
	fmt.Println("=== SUNAT Document Validation Examples ===\n")

	// Create validation client with master credentials from environment variables
	// Set these environment variables: SUNAT_MASTER_RUC, SUNAT_MASTER_USERNAME, SUNAT_MASTER_PASSWORD
	masterRUC := os.Getenv("SUNAT_MASTER_RUC")
	masterUsername := os.Getenv("SUNAT_MASTER_USERNAME")
	masterPassword := os.Getenv("SUNAT_MASTER_PASSWORD")

	if masterRUC == "" || masterUsername == "" || masterPassword == "" {
		log.Fatal("Please set SUNAT_MASTER_RUC, SUNAT_MASTER_USERNAME, and SUNAT_MASTER_PASSWORD environment variables")
	}

	validator := sunatlib.NewValidationClient(masterRUC, masterUsername, masterPassword)

	// Example 2: Validate an invoice
	fmt.Println("1. Validating an invoice...")
	invoiceResult, err := validator.ValidateInvoice(
		"20123456789",     // Issuer RUC
		"F001",            // Series number
		"00000001",        // Document number
		"2024-01-15",      // Issue date (YYYY-MM-DD)
		1250.50,           // Total amount
	)
	if err != nil {
		log.Printf("Error validating invoice: %v", err)
	} else {
		printValidationResult("Invoice", invoiceResult)
	}

	// Example 3: Validate a receipt
	fmt.Println("2. Validating a receipt...")
	receiptResult, err := validator.ValidateReceipt(
		"20123456789",     // Issuer RUC
		"B001",            // Series number
		"00000001",        // Document number
		"2024-01-15",      // Issue date (YYYY-MM-DD)
		85.50,             // Total amount
	)
	if err != nil {
		log.Printf("Error validating receipt: %v", err)
	} else {
		printValidationResult("Receipt", receiptResult)
	}

	// Example 4: Custom validation with full parameters
	fmt.Println("3. Custom validation with full parameters...")
	customParams := &sunatlib.ValidationParams{
		IssuerRUC:           "20123456789",
		DocumentType:        "01", // Invoice
		SeriesNumber:        "F001",
		DocumentNumber:      "00000002",
		RecipientDocType:    "6", // RUC
		RecipientDocNumber:  "20987654321",
		IssueDate:           "2024-01-15",
		TotalAmount:         2500.75,
		AuthorizationNumber: "",
	}

	customResult, err := validator.ValidateDocument(customParams)
	if err != nil {
		log.Printf("Error in custom validation: %v", err)
	} else {
		printValidationResult("Custom Document", customResult)
	}

	// Example 5: Using different master credentials
	fmt.Println("4. Using different master credentials...")
	// You can use different master credentials for different validation scenarios
	customValidator := sunatlib.NewValidationClient(
		os.Getenv("SUNAT_ALT_RUC"),      // Alternative master RUC
		os.Getenv("SUNAT_ALT_USERNAME"), // Alternative master username
		os.Getenv("SUNAT_ALT_PASSWORD"), // Alternative master password
	)

	if os.Getenv("SUNAT_ALT_RUC") != "" {
		customCredResult, err := customValidator.ValidateInvoice(
			"20123456789",
			"F001",
			"00000003",
			"2024-01-15",
			3750.25,
		)
		if err != nil {
			log.Printf("Error with alternative credentials: %v", err)
		} else {
			printValidationResult("Invoice (Alternative Credentials)", customCredResult)
		}
	} else {
		fmt.Println("   Skipping alternative credentials example (SUNAT_ALT_* not set)")
	}
}

func printValidationResult(docType string, result *sunatlib.ValidationResult) {
	fmt.Printf("   %s Validation Result:\n", docType)
	fmt.Printf("   - Success: %t\n", result.Success)
	fmt.Printf("   - Is Valid: %t\n", result.IsValid)
	fmt.Printf("   - State: %s\n", result.State)
	fmt.Printf("   - Status Code: %s\n", result.StatusCode)
	fmt.Printf("   - Status Message: %s\n", result.StatusMessage)
	fmt.Printf("   - Error Details: %s\n", result.ErrorDetails)
	fmt.Println()
}