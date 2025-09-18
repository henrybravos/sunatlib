package main

import (
	"fmt"
	"log"
	"time"

	"github.com/henrybravos/sunatlib"
)

func main() {
	// Example: Sending voided documents communication to SUNAT
	fmt.Println("=== SUNAT Voided Documents (Comunicación de Baja) Example ===")

	// Configure SUNAT client
	client := sunatlib.NewVoidedDocumentsClient(
		"20123456789",  // Your RUC number
		"MODDATOS",     // SOL username
		"moddatos",     // SOL password
		"https://e-beta.sunat.gob.pe/ol-ti-itcpfegem-beta/billService", // Beta endpoint
	)
	defer client.Cleanup()

	// Configure certificate from PFX (optional - for signing)
	err := client.SetCertificateFromPFX("certificate.pfx", "your_password", "/tmp/certs")
	if err != nil {
		fmt.Printf("Warning: Could not set certificate: %v\n", err)
		fmt.Println("Continuing without digital signature...")
	}

	// Create voided documents request
	now := time.Now()
	referenceDate := now.AddDate(0, 0, -1) // Yesterday's documents

	request := &sunatlib.VoidedDocumentsRequest{
		RUC:           "20123456789",
		CompanyName:   "MI EMPRESA S.A.C.",
		SeriesNumber:  sunatlib.GenerateVoidedDocumentsSeries(referenceDate, 1), // RA-YYYYMMDD-001
		IssueDate:     now,
		ReferenceDate: referenceDate,
		Description:   "Comunicación de baja de documentos",
		Documents: []sunatlib.VoidedDocument{
			{
				DocumentTypeCode: "01",     // Invoice
				DocumentSeries:   "F001",   // Series
				DocumentNumber:   "000123", // Document number
				VoidedReason:     "Error en datos del cliente",
			},
			{
				DocumentTypeCode: "03",     // Receipt
				DocumentSeries:   "B001",   // Series
				DocumentNumber:   "000456", // Document number
				VoidedReason:     "Duplicado por error del sistema",
			},
		},
	}

	// Send voided documents communication
	fmt.Printf("Sending voided documents communication: %s\n", request.SeriesNumber)
	response, err := client.SendVoidedDocuments(request)
	if err != nil {
		log.Fatal("Error sending voided documents:", err)
	}

	// Check response
	if response.Success {
		fmt.Printf("✅ Voided documents communication sent successfully\n")
		fmt.Printf("📄 Message: %s\n", response.Message)
		fmt.Printf("🎫 Ticket: %s\n", response.Ticket)

		// Wait a moment and check status
		fmt.Println("\n⏳ Waiting 5 seconds before checking status...")
		time.Sleep(5 * time.Second)

		// Check status using ticket
		fmt.Printf("Checking status for ticket: %s\n", response.Ticket)
		statusResponse, err := client.GetVoidedDocumentsStatus(response.Ticket)
		if err != nil {
			log.Printf("Error checking status: %v", err)
		} else {
			if statusResponse.Success {
				fmt.Printf("✅ Status check successful\n")
				fmt.Printf("📄 Status message: %s\n", statusResponse.Message)

				// Save CDR if available
				if statusResponse.ApplicationResponse != nil {
					err = statusResponse.SaveApplicationResponse("voided_cdr.zip")
					if err != nil {
						log.Printf("Error saving CDR: %v", err)
					} else {
						fmt.Println("💾 CDR saved as voided_cdr.zip")
					}
				}
			} else {
				fmt.Printf("⏳ Status: %s\n", statusResponse.Message)
				fmt.Println("Note: Processing may still be in progress. Check again later.")
			}
		}
	} else {
		fmt.Printf("❌ Error: %s\n", response.Message)
	}

	fmt.Println("\n=== Example completed ===")
}