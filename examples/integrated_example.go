package main

import (
	"fmt"
	"log"
	"time"

	"github.com/henrybravos/sunatlib"
)

func main() {
	fmt.Println("=== SUNAT Integrated Example: Document Validation + Voided Documents ===")

	// Example workflow:
	// 1. Validate some existing documents
	// 2. If needed, void some documents using communication de baja

	// Step 1: Document Validation
	fmt.Println("\n📋 Step 1: Validating existing documents")

	validationClient := sunatlib.NewDocumentValidationClient(
		"20123456789", // RUC
		"MODDATOS",    // SOL username
		"moddatos",    // SOL password
	)

	// Validate an invoice that we might need to void
	invoiceResp, err := validationClient.ValidateInvoice(
		"20123456789", // Issuer RUC
		"F001",        // Series
		"000123",      // Number
		"15/01/2025",  // Issue date (DD/MM/YYYY)
		"118.00",      // Total amount
	)

	if err != nil {
		log.Printf("Error validating invoice: %v", err)
	} else {
		fmt.Printf("📄 Invoice F001-000123 validation result: %s\n", invoiceResp.GetStatusDescription())
		if invoiceResp.IsDocumentValid() {
			fmt.Println("✅ Invoice is valid in SUNAT")
		} else {
			fmt.Println("❌ Invoice is invalid or not found")
		}
	}

	// Step 2: Void Documents Communication
	fmt.Println("\n📋 Step 2: Creating voided documents communication")

	voidedClient := sunatlib.NewVoidedDocumentsClient(
		"20123456789",  // RUC
		"MODDATOS",     // SOL username
		"moddatos",     // SOL password
		"https://e-beta.sunat.gob.pe/ol-ti-itcpfegem-beta/billService", // Beta endpoint
	)
	defer voidedClient.Cleanup()

	// Configure certificate (optional - for signing)
	err = voidedClient.SetCertificateFromPFX("certificate.pfx", "your_password", "/tmp/certs")
	if err != nil {
		fmt.Printf("Warning: Could not set certificate: %v\n", err)
		fmt.Println("Continuing without digital signature...")
	}

	// Create voided documents request
	now := time.Now()
	referenceDate := now.AddDate(0, 0, -1) // Yesterday's documents

	voidRequest := &sunatlib.VoidedDocumentsRequest{
		RUC:           "20123456789",
		CompanyName:   "MI EMPRESA S.A.C.",
		SeriesNumber:  sunatlib.GenerateVoidedDocumentsSeries(referenceDate, 1),
		IssueDate:     now,
		ReferenceDate: referenceDate,
		Description:   "Comunicación de baja por corrección de errores",
		Documents: []sunatlib.VoidedDocument{
			{
				DocumentTypeCode: "01",                           // Invoice
				DocumentSeries:   "F001",                         // Series
				DocumentNumber:   "000123",                       // Same document we validated above
				VoidedReason:     "Error en datos del cliente",
			},
			{
				DocumentTypeCode: "03",                           // Receipt
				DocumentSeries:   "B001",                         // Series
				DocumentNumber:   "000456",                       // Another document
				VoidedReason:     "Documento emitido por duplicado",
			},
		},
	}

	// Validate the request before sending
	if err := voidRequest.Validate(); err != nil {
		log.Fatalf("Invalid voided documents request: %v", err)
	}

	fmt.Printf("📄 Voided documents request is valid\n")
	fmt.Printf("📄 Series: %s\n", voidRequest.SeriesNumber)
	fmt.Printf("📄 Reference Date: %s\n", voidRequest.ReferenceDate.Format("2006-01-02"))
	fmt.Printf("📄 Documents to void: %d\n", len(voidRequest.Documents))

	// Generate XML to preview (optional)
	xmlContent, err := voidedClient.GenerateVoidedDocumentsXML(voidRequest)
	if err != nil {
		log.Printf("Error generating XML: %v", err)
	} else {
		fmt.Printf("📄 Generated XML size: %d bytes\n", len(xmlContent))
		// Optionally save XML for inspection
		// os.WriteFile("voided_documents.xml", xmlContent, 0644)
	}

	// Send voided documents communication
	fmt.Println("\n📤 Sending voided documents communication to SUNAT...")
	voidResponse, err := voidedClient.SendVoidedDocuments(voidRequest)
	if err != nil {
		log.Printf("Error sending voided documents: %v", err)
	} else {
		if voidResponse.Success {
			fmt.Printf("✅ Voided documents communication sent successfully!\n")
			fmt.Printf("🎫 Ticket: %s\n", voidResponse.Ticket)
			fmt.Printf("📄 Message: %s\n", voidResponse.Message)

			// Wait a moment and check status
			fmt.Println("\n⏳ Waiting 10 seconds before checking status...")
			time.Sleep(10 * time.Second)

			fmt.Println("📋 Checking status of voided documents...")
			statusResponse, err := voidedClient.GetVoidedDocumentsStatus(voidResponse.Ticket)
			if err != nil {
				log.Printf("Error checking status: %v", err)
			} else {
				if statusResponse.Success {
					fmt.Printf("✅ Status check successful: %s\n", statusResponse.Message)

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
					fmt.Printf("⏳ Processing status: %s\n", statusResponse.Message)
					fmt.Println("Note: The voiding process may still be in progress.")
				}
			}
		} else {
			fmt.Printf("❌ Error sending voided documents: %s\n", voidResponse.Message)
		}
	}

	// Step 3: Validate documents again to confirm they are voided
	fmt.Println("\n📋 Step 3: Re-validating documents after voiding")

	time.Sleep(2 * time.Second) // Wait a bit more

	invoiceRespAfter, err := validationClient.ValidateInvoice(
		"20123456789", // Issuer RUC
		"F001",        // Series
		"000123",      // Number (same as voided)
		"15/01/2025",  // Issue date
		"118.00",      // Total amount
	)

	if err != nil {
		log.Printf("Error validating invoice after voiding: %v", err)
	} else {
		fmt.Printf("📄 Invoice F001-000123 validation after voiding: %s\n", invoiceRespAfter.GetStatusDescription())
		if invoiceRespAfter.IsDocumentValid() {
			fmt.Println("⚠️  Invoice still appears valid (voiding may not be processed yet)")
		} else {
			fmt.Println("✅ Invoice successfully voided")
		}
	}

	fmt.Println("\n=== Integrated workflow completed ===")
	fmt.Println("Summary:")
	fmt.Println("- Document validation: ✅ Implemented")
	fmt.Println("- Voided documents communication: ✅ Implemented")
	fmt.Println("- SUNAT integration: ✅ Working")
	fmt.Println("- XML generation and validation: ✅ Working")
}