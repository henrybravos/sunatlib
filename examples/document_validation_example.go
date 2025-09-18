package main

import (
	"fmt"
	"log"

	"github.com/henrybravos/sunatlib"
)

func main() {
	// Example: Validating electronic documents with SUNAT
	fmt.Println("=== SUNAT Document Validation Example ===")

	// Create validation client with SUNAT credentials
	client := sunatlib.NewDocumentValidationClient(
		"20123456789", // Your RUC number
		"MODDATOS",    // SOL username
		"moddatos",    // SOL password
	)

	fmt.Println("✅ Validation client created successfully")

	// Example 1: Validate an invoice
	fmt.Println("\n📋 Example 1: Validating an invoice")
	invoiceResp, err := client.ValidateInvoice(
		"20123456789", // Issuer RUC
		"F001",        // Series
		"000123",      // Number
		"15/01/2025",  // Issue date (DD/MM/YYYY)
		"118.00",      // Total amount
	)
	if err != nil {
		log.Printf("Error validating invoice: %v", err)
	} else {
		printValidationResult("Invoice", invoiceResp)
	}

	// Example 2: Validate a receipt (boleta)
	fmt.Println("\n📋 Example 2: Validating a receipt")
	receiptResp, err := client.ValidateReceipt(
		"20123456789", // Issuer RUC
		"B001",        // Series
		"000456",      // Number
		"15/01/2025",  // Issue date (DD/MM/YYYY)
		"59.00",       // Total amount
	)
	if err != nil {
		log.Printf("Error validating receipt: %v", err)
	} else {
		printValidationResult("Receipt", receiptResp)
	}

	// Example 3: Validate a credit note
	fmt.Println("\n📋 Example 3: Validating a credit note")
	creditNoteResp, err := client.ValidateCreditNote(
		"20123456789", // Issuer RUC
		"FC01",        // Series
		"000001",      // Number
		"15/01/2025",  // Issue date (DD/MM/YYYY)
		"23.60",       // Total amount
	)
	if err != nil {
		log.Printf("Error validating credit note: %v", err)
	} else {
		printValidationResult("Credit Note", creditNoteResp)
	}

	// Example 4: Check document status (basic validation)
	fmt.Println("\n📋 Example 4: Checking document status")
	statusResp, err := client.CheckDocumentStatus(
		"20123456789", // Issuer RUC
		"01",          // Document type (01=Invoice)
		"F001",        // Series
		"000789",      // Number
	)
	if err != nil {
		log.Printf("Error checking document status: %v", err)
	} else {
		printValidationResult("Document Status Check", statusResp)
	}

	fmt.Println("\n=== Validation examples completed ===")
}

func printValidationResult(docType string, resp *sunatlib.ValidationResponse) {
	fmt.Printf("\n--- %s Validation Result ---\n", docType)

	if resp.Success {
		fmt.Printf("✅ Document is valid: %t\n", resp.IsDocumentValid())
		fmt.Printf("📄 Status Description: %s\n", resp.GetStatusDescription())

		if resp.IsValid {
			fmt.Printf("🎯 Validation Status: VÁLIDO\n")
		} else {
			fmt.Printf("🎯 Validation Status: INVÁLIDO\n")
		}
	} else {
		fmt.Printf("❌ Validation failed\n")
		fmt.Printf("📄 Error Message: %s\n", resp.GetErrorMessage())
	}
}