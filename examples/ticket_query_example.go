// Package main demonstrates how to query voided document communication tickets
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/henrybravos/sunatlib"
)

func runTicketQueryExamples() {
	// Example 1: Basic ticket query
	basicTicketQuery()

	// Example 2: Waiting for ticket processing with timeout
	waitForTicketProcessing()

	// Example 3: Batch query multiple tickets
	batchQueryExample()

	// Example 4: Complete workflow - send voided documents and query status
	completeWorkflowExample()
}

// basicTicketQuery demonstrates basic ticket status query
func basicTicketQuery() {
	fmt.Println("=== Basic Ticket Query Example ===")

	// Create client for production environment
	client := sunatlib.NewVoidedDocumentsClient(
		"20123456789", // Your RUC
		"USUARIO",     // Your SOL username
		"PASSWORD",    // Your SOL password
	)
	defer client.Cleanup()

	// Query ticket status
	ticket := "12345678901234567890" // Replace with actual ticket
	response, err := client.QueryVoidedDocumentsTicket(ticket)
	if err != nil {
		log.Printf("Error querying ticket: %v", err)
		return
	}

	fmt.Printf("Ticket: %s\n", response.Ticket)
	fmt.Printf("Success: %t\n", response.Success)
	fmt.Printf("Message: %s\n", response.Message)
	fmt.Printf("Status Code: %s\n", response.StatusCode)
	fmt.Printf("Status Description: %s\n", response.StatusDescription)

	// Check status using convenience methods
	if response.IsSuccessful() {
		fmt.Println("✅ Comunicación de baja procesada exitosamente")

		// Get CDR data and let user decide where to save it
		if response.HasApplicationResponse() {
			cdrData := response.GetApplicationResponse()
			fmt.Printf("📁 CDR disponible (%d bytes) - el usuario puede guardarlo donde desee\n", len(cdrData))
			// Example: os.WriteFile("user_chosen_path/cdr_baja.zip", cdrData, 0644)
		}
	} else if response.IsInProgress() {
		fmt.Println("⏳ Comunicación de baja en proceso...")
	} else if response.HasErrors() {
		fmt.Println("❌ Comunicación de baja procesada con errores")

		// Get error details if available
		if response.HasApplicationResponse() {
			errorData := response.GetApplicationResponse()
			fmt.Printf("📁 Detalles del error disponibles (%d bytes) - el usuario puede analizarlos\n", len(errorData))
			// Example: os.WriteFile("user_chosen_path/error_details.zip", errorData, 0644)
		}
	}

	fmt.Println()
}

// waitForTicketProcessing demonstrates waiting for ticket processing with timeout
func waitForTicketProcessing() {
	fmt.Println("=== Wait for Ticket Processing Example ===")

	// Create client for beta testing environment
	client := sunatlib.NewVoidedDocumentsClientBeta(
		"20123456789", // Test RUC
		"MODDATOS",    // Test username
		"moddatos",    // Test password
	)
	defer client.Cleanup()

	ticket := "12345678901234567890" // Replace with actual ticket

	// Wait up to 10 minutes, checking every 30 seconds
	maxWaitTime := 10 * time.Minute
	pollInterval := 30 * time.Second

	fmt.Printf("Esperando procesamiento del ticket %s...\n", ticket)
	fmt.Printf("Tiempo máximo de espera: %v\n", maxWaitTime)
	fmt.Printf("Intervalo de consulta: %v\n", pollInterval)

	response, err := client.WaitForTicketProcessing(ticket, maxWaitTime, pollInterval)
	if err != nil {
		log.Printf("Error waiting for ticket processing: %v", err)
		return
	}

	fmt.Printf("Resultado final: %s\n", response.Message)

	if response.IsSuccessful() {
		fmt.Println("🎉 Procesamiento completado exitosamente")
	} else if response.HasErrors() {
		fmt.Println("⚠️ Procesamiento completado con errores")
	} else if response.IsInProgress() {
		fmt.Println("⏰ Timeout - el ticket aún está en proceso")
	}

	fmt.Println()
}

// batchQueryExample demonstrates querying multiple tickets at once
func batchQueryExample() {
	fmt.Println("=== Batch Query Example ===")

	client := sunatlib.NewVoidedDocumentsClient(
		"20123456789",
		"USUARIO",
		"PASSWORD",
	)
	defer client.Cleanup()

	// Multiple tickets to query
	tickets := []string{
		"12345678901234567890",
		"12345678901234567891",
		"12345678901234567892",
		"12345678901234567893",
	}

	fmt.Printf("Consultando %d tickets...\n", len(tickets))

	responses, err := client.BatchQueryTickets(tickets)
	if err != nil {
		log.Printf("Error in batch query: %v", err)
		return
	}

	// Process results
	successCount := 0
	errorCount := 0
	inProgressCount := 0

	for i, response := range responses {
		fmt.Printf("Ticket %d: %s\n", i+1, response.Ticket)
		fmt.Printf("  Status: %s\n", response.GetTicketStatusDescription())

		if response.IsSuccessful() {
			successCount++
			fmt.Printf("  ✅ Procesado exitosamente\n")
		} else if response.HasErrors() {
			errorCount++
			fmt.Printf("  ❌ Procesado con errores\n")
		} else if response.IsInProgress() {
			inProgressCount++
			fmt.Printf("  ⏳ En proceso\n")
		} else {
			fmt.Printf("  ❓ Estado desconocido: %s\n", response.Message)
		}
	}

	fmt.Printf("\nResumen:\n")
	fmt.Printf("  Exitosos: %d\n", successCount)
	fmt.Printf("  Con errores: %d\n", errorCount)
	fmt.Printf("  En proceso: %d\n", inProgressCount)

	fmt.Println()
}

// completeWorkflowExample demonstrates complete workflow from sending to querying
func completeWorkflowExample() {
	fmt.Println("=== Complete Workflow Example ===")

	// Use beta environment for testing
	client := sunatlib.NewVoidedDocumentsClientBeta(
		"20123456789",
		"MODDATOS",
		"moddatos",
	)
	defer client.Cleanup()

	// Configure certificate
	err := client.SetCertificateFromPFX("certificate.pfx", "password", "/tmp/certs")
	if err != nil {
		log.Printf("Certificate error (skipping for demo): %v", err)
		// In real usage, you would return here
	}

	// Step 1: Create and send voided documents communication
	now := time.Now()
	referenceDate := now.AddDate(0, 0, -1) // Yesterday's documents

	request := &sunatlib.VoidedDocumentsRequest{
		RUC:           "20123456789",
		CompanyName:   "MI EMPRESA S.A.C.",
		SeriesNumber:  sunatlib.GenerateVoidedDocumentsSeries(referenceDate, 1),
		IssueDate:     now,
		ReferenceDate: referenceDate,
		Description:   "Comunicación de baja de documentos",
		Documents: []sunatlib.VoidedDocument{
			{
				DocumentTypeCode: "01",
				DocumentSeries:   "F001",
				DocumentNumber:   "000123",
				VoidedReason:     "Error en datos del cliente",
			},
			{
				DocumentTypeCode: "03",
				DocumentSeries:   "B001",
				DocumentNumber:   "000456",
				VoidedReason:     "Duplicado por error",
			},
		},
	}

	fmt.Printf("Enviando comunicación de baja: %s\n", request.SeriesNumber)

	voidResponse, err := client.SendVoidedDocuments(request)
	if err != nil {
		log.Printf("Error sending voided documents: %v", err)
		return
	}

	if !voidResponse.Success {
		fmt.Printf("❌ Error al enviar: %s\n", voidResponse.Message)
		return
	}

	fmt.Printf("✅ Comunicación enviada. Ticket: %s\n", voidResponse.Ticket)

	// Step 2: Wait for processing
	fmt.Println("Esperando procesamiento...")

	finalResponse, err := client.WaitForTicketProcessing(
		voidResponse.Ticket,
		5*time.Minute, // Wait up to 5 minutes
		15*time.Second, // Check every 15 seconds
	)
	if err != nil {
		log.Printf("Error waiting for processing: %v", err)
		return
	}

	// Step 3: Handle final result
	fmt.Printf("Resultado final: %s\n", finalResponse.Message)

	switch {
	case finalResponse.IsSuccessful():
		fmt.Println("🎉 Comunicación de baja procesada exitosamente")
		if finalResponse.HasApplicationResponse() {
			cdrData := finalResponse.GetApplicationResponse()
			fmt.Printf("📁 CDR disponible (%d bytes) para guardar\n", len(cdrData))
			// User can save: os.WriteFile("comunicacion_baja_cdr.zip", cdrData, 0644)
		}

	case finalResponse.HasErrors():
		fmt.Println("⚠️ Comunicación procesada con errores")
		if finalResponse.HasApplicationResponse() {
			errorData := finalResponse.GetApplicationResponse()
			fmt.Printf("📁 Detalles de errores disponibles (%d bytes)\n", len(errorData))
			// User can save: os.WriteFile("comunicacion_baja_errors.zip", errorData, 0644)
		}

	case finalResponse.IsInProgress():
		fmt.Println("⏰ Timeout - la comunicación aún está en proceso")
		fmt.Println("Puede consultar el estado más tarde usando el ticket:", finalResponse.Ticket)

	default:
		fmt.Printf("❓ Estado inesperado: %s\n", finalResponse.Message)
	}

	fmt.Println()
}

// Example of how to use in a real application
func realWorldExample() {
	fmt.Println("=== Real World Usage Pattern ===")

	// Production configuration
	client := sunatlib.NewVoidedDocumentsClient(
		"20123456789", // Your real RUC
		"USUARIO",     // Your real SOL username
		"PASSWORD",    // Your real SOL password
	)
	defer client.Cleanup()

	// In a real application, you might store tickets in a database
	// and check their status periodically using a cron job or background worker

	tickets := []string{
		// These would come from your database
		"12345678901234567890",
		"12345678901234567891",
	}

	for _, ticket := range tickets {
		response, err := client.QueryVoidedDocumentsTicket(ticket)
		if err != nil {
			log.Printf("Error querying ticket %s: %v", ticket, err)
			continue
		}

		// Update database based on response
		if response.IsSuccessful() {
			// Mark as completed in database
			fmt.Printf("Ticket %s: COMPLETED\n", ticket)

			// Get CDR data for user to handle
			if response.HasApplicationResponse() {
				cdrData := response.GetApplicationResponse()
				fmt.Printf("  CDR available (%d bytes)\n", len(cdrData))
				// User can save: os.WriteFile(fmt.Sprintf("cdr_%s.zip", ticket), cdrData, 0644)
			}

		} else if response.HasErrors() {
			// Mark as failed in database
			fmt.Printf("Ticket %s: FAILED\n", ticket)

			// Get error details for user to handle
			if response.HasApplicationResponse() {
				errorData := response.GetApplicationResponse()
				fmt.Printf("  Error details available (%d bytes)\n", len(errorData))
				// User can save: os.WriteFile(fmt.Sprintf("error_%s.zip", ticket), errorData, 0644)
			}

		} else if response.IsInProgress() {
			// Keep checking later
			fmt.Printf("Ticket %s: PENDING\n", ticket)

		} else {
			// Handle unexpected status
			fmt.Printf("Ticket %s: UNKNOWN - %s\n", ticket, response.Message)
		}
	}
}