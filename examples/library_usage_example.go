// Package main demonstrates the correct usage of the library following best practices
// The library provides data, users decide what to do with it
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/henrybravos/sunatlib"
)

func demonstrateLibraryUsage() {
	fmt.Println("=== Ejemplo de Uso Correcto de la Librería ===")
	fmt.Println("La librería proporciona datos, el usuario decide qué hacer con ellos\n")

	// Create client
	client := sunatlib.NewVoidedDocumentsClientBeta(
		"20123456789",
		"MODDATOS",
		"moddatos",
	)
	defer client.Cleanup()

	// Example 1: Query ticket and let user handle the response data
	ticket := "12345678901234567890"
	response, err := client.QueryVoidedDocumentsTicket(ticket)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Ticket: %s\n", response.Ticket)
	fmt.Printf("Estado: %s\n", response.GetTicketStatusDescription())

	// The library provides the data, user decides what to do
	if response.IsSuccessful() && response.HasApplicationResponse() {
		cdrData := response.GetApplicationResponse()

		// User can choose to:
		// 1. Save to a specific location
		userDirectory := "/home/user/documents/sunat/"
		os.MkdirAll(userDirectory, 0755)
		os.WriteFile(userDirectory+"cdr_"+ticket+".zip", cdrData, 0644)

		// 2. Process the data directly
		fmt.Printf("CDR size: %d bytes\n", len(cdrData))

		// 3. Store in database, send via email, upload to cloud, etc.
		// storeInDatabase(ticket, cdrData)
		// sendViaEmail(cdrData)
		// uploadToCloud(cdrData)

		fmt.Println("✅ CDR data obtained and handled by user")
	}

	// Example 2: Batch processing with user-controlled data handling
	fmt.Println("\n=== Procesamiento por Lotes ===")

	tickets := []string{
		"12345678901234567890",
		"12345678901234567891",
		"12345678901234567892",
	}

	responses, err := client.BatchQueryTickets(tickets)
	if err != nil {
		log.Printf("Error in batch query: %v", err)
		return
	}

	// Process each response according to user needs
	for i, resp := range responses {
		fmt.Printf("Ticket %d (%s): %s\n", i+1, resp.Ticket, resp.GetTicketStatusDescription())

		if resp.HasApplicationResponse() {
			data := resp.GetApplicationResponse()

			// User decides how to handle each response based on their business logic
			if resp.IsSuccessful() {
				// Save successful CDRs to one directory
				savePath := fmt.Sprintf("/user/successful/cdr_%s.zip", resp.Ticket)
				os.WriteFile(savePath, data, 0644)
				fmt.Printf("  ✅ CDR saved to %s\n", savePath)

			} else if resp.HasErrors() {
				// Save error details to another directory
				errorPath := fmt.Sprintf("/user/errors/error_%s.zip", resp.Ticket)
				os.WriteFile(errorPath, data, 0644)
				fmt.Printf("  ❌ Error details saved to %s\n", errorPath)

				// User might also want to:
				// - Log to their logging system
				// - Send alerts
				// - Update database status
			}
		}
	}

	// Example 3: Real-world integration pattern
	fmt.Println("\n=== Patrón de Integración del Mundo Real ===")

	// In a real application, you might have a service that:
	processTicketsForBusiness(client, []string{"ticket1", "ticket2"})
}

// Example of how a business might integrate the library
func processTicketsForBusiness(client *sunatlib.SUNATClient, tickets []string) {
	for _, ticket := range tickets {
		response, err := client.QueryVoidedDocumentsTicket(ticket)
		if err != nil {
			// Handle error according to business needs
			logError("TICKET_QUERY_ERROR", ticket, err)
			continue
		}

		// Update business database
		updateTicketStatus(ticket, response)

		// Handle response data based on business requirements
		if response.HasApplicationResponse() {
			data := response.GetApplicationResponse()

			switch {
			case response.IsSuccessful():
				// Business logic for successful processing
				archiveCDR(ticket, data)
				notifySuccess(ticket)

			case response.HasErrors():
				// Business logic for error handling
				logErrorDetails(ticket, data)
				notifyError(ticket, response.Message)

			case response.IsInProgress():
				// Schedule for retry later
				scheduleRetry(ticket)
			}
		}
	}
}

// Mock business functions to demonstrate integration patterns
func logError(errorType, ticket string, err error) {
	fmt.Printf("[ERROR] %s for ticket %s: %v\n", errorType, ticket, err)
}

func updateTicketStatus(ticket string, response *sunatlib.TicketStatusResponse) {
	fmt.Printf("[DB] Updating ticket %s status to: %s\n", ticket, response.GetTicketStatusDescription())
}

func archiveCDR(ticket string, data []byte) {
	// Business-specific storage location
	archivePath := fmt.Sprintf("/company/archive/cdr/%s.zip", ticket)
	os.WriteFile(archivePath, data, 0644)
	fmt.Printf("[ARCHIVE] CDR for ticket %s archived\n", ticket)
}

func notifySuccess(ticket string) {
	fmt.Printf("[NOTIFICATION] Success notification sent for ticket %s\n", ticket)
}

func logErrorDetails(ticket string, data []byte) {
	// Extract and log error details for business analysis
	errorPath := fmt.Sprintf("/company/errors/%s.zip", ticket)
	os.WriteFile(errorPath, data, 0644)
	fmt.Printf("[ERROR_LOG] Error details for ticket %s logged\n", ticket)
}

func notifyError(ticket, message string) {
	fmt.Printf("[NOTIFICATION] Error notification sent for ticket %s: %s\n", ticket, message)
}

func scheduleRetry(ticket string) {
	fmt.Printf("[SCHEDULER] Ticket %s scheduled for retry\n", ticket)
}

// Demonstrate flexible timeout handling
func demonstrateTimeoutHandling() {
	fmt.Println("\n=== Manejo Flexible de Timeouts ===")

	client := sunatlib.NewVoidedDocumentsClient("20123456789", "USER", "PASS")
	defer client.Cleanup()

	ticket := "12345678901234567890"

	// Different timeout strategies based on business needs

	// 1. Quick check (for real-time applications)
	response, err := client.QueryVoidedDocumentsTicket(ticket)
	if err == nil && response.IsInProgress() {
		fmt.Println("Ticket still processing, will check again later")
		return
	}

	// 2. Medium wait (for batch processing)
	finalResponse, err := client.WaitForTicketProcessing(
		ticket,
		5*time.Minute,  // Wait up to 5 minutes
		30*time.Second, // Check every 30 seconds
	)
	if err == nil && finalResponse.IsProcessed() {
		fmt.Println("Ticket processed within 5 minutes")
		// Handle the response data as needed
		if finalResponse.HasApplicationResponse() {
			data := finalResponse.GetApplicationResponse()
			// User handles the data according to their needs
			fmt.Printf("Response data available: %d bytes\n", len(data))
		}
	}

	// 3. Long wait (for overnight processing)
	// client.WaitForTicketProcessing(ticket, 2*time.Hour, 2*time.Minute)
}