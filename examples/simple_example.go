package main

import (
	"fmt"
	"log"
	"os"

	"github.com/henrybravos/sunatlib"
)

func main() {
	// Example usage of SUNATLib
	err := sendInvoiceExample()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func sendInvoiceExample() error {
	// Create SUNAT client with your credentials
	client := sunatlib.NewSUNATClient(
		"20123456789",     // Your RUC number
		"MODDATOS",        // SOL username
		"moddatos",        // SOL password  
		"https://e-beta.sunat.gob.pe/ol-ti-itcpfegem-beta/billService", // Beta endpoint
	)
	defer client.Cleanup()

	// Configure certificate from PFX file
	err := client.SetCertificateFromPFX(
		"./certificate.pfx",    // Path to your certificate
		"your_cert_password",   // Certificate password
		"/tmp/sunatlib_certs",  // Temporary directory
	)
	if err != nil {
		return fmt.Errorf("failed to configure certificate: %w", err)
	}

	// Read invoice XML file
	xmlContent, err := os.ReadFile("./invoice.xml")
	if err != nil {
		return fmt.Errorf("failed to read XML file: %w", err)
	}

	// Sign and send invoice to SUNAT
	response, err := client.SignAndSendInvoice(
		xmlContent,      // XML content
		"01",           // Document type (invoice)
		"F001-00000001", // Series-number
	)
	if err != nil {
		return fmt.Errorf("failed to send invoice: %w", err)
	}

	// Process response
	fmt.Printf("✅ Success: %t\n", response.Success)
	fmt.Printf("📝 Message: %s\n", response.Message)

	if response.Success && response.ApplicationResponse != nil {
		err = response.SaveApplicationResponse("cdr_response.zip")
		if err != nil {
			log.Printf("⚠️ Failed to save CDR: %v", err)
		} else {
			fmt.Println("💾 CDR saved as: cdr_response.zip")
		}
	}

	return nil
}