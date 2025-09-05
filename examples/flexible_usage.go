package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/henrybravos/sunatlib"
)

func main() {
	// Example showing flexible usage patterns
	fmt.Println("🔧 Flexible usage examples:")
	
	// Example 1: Sign only (no send)
	signOnlyExample()
	
	// Example 2: Batch processing
	batchProcessingExample()
	
	// Example 3: Sign now, send later
	signNowSendLaterExample()
}

// Example 1: Sign XML documents without sending to SUNAT
func signOnlyExample() {
	fmt.Println("\n1️⃣ Sign-only example (no SUNAT transmission):")
	
	client := sunatlib.NewSUNATClient("", "", "", "") // No SUNAT credentials needed
	defer client.Cleanup()
	
	// Configure certificate for signing only
	err := client.SetCertificateFromPFX(
		"./certificate.pfx",
		"password",
		"/tmp/sunatlib_sign_only",
	)
	if err != nil {
		fmt.Printf("⚠️ Certificate error (expected): %v\n", err)
		return
	}
	
	// Read and sign multiple documents
	documents := []string{"invoice1.xml", "invoice2.xml", "creditnote1.xml"}
	
	for _, docName := range documents {
		fmt.Printf("🔏 Processing %s...\n", docName)
		
		// In real usage, read actual file
		xmlContent := []byte(`<Invoice>Sample content for ` + docName + `</Invoice>`)
		
		signedXML, err := client.SignXML(xmlContent)
		if err != nil {
			fmt.Printf("❌ Failed to sign %s: %v\n", docName, err)
			continue
		}
		
		// Save signed document
		outputName := "signed_" + docName
		err = os.WriteFile(outputName, signedXML, 0644)
		if err != nil {
			fmt.Printf("⚠️ Failed to save %s\n", outputName)
		} else {
			fmt.Printf("✅ Signed and saved: %s\n", outputName)
		}
	}
}

// Example 2: Batch processing with error handling
func batchProcessingExample() {
	fmt.Println("\n2️⃣ Batch processing example:")
	
	client := sunatlib.NewSUNATClient(
		"20123456789",
		"MODDATOS", 
		"moddatos",
		"https://e-beta.sunat.gob.pe/ol-ti-itcpfegem-beta/billService",
	)
	defer client.Cleanup()
	
	// Simulate batch of documents
	invoices := []struct {
		filename string
		docType  string
		series   string
	}{
		{"invoice001.xml", "01", "F001-00000001"},
		{"invoice002.xml", "01", "F001-00000002"}, 
		{"invoice003.xml", "01", "F001-00000003"},
	}
	
	results := make([]string, 0, len(invoices))
	
	for _, inv := range invoices {
		fmt.Printf("📄 Processing %s (%s-%s)...\n", inv.filename, inv.docType, inv.series)
		
		// Read XML (simulate)
		xmlContent := []byte(`<Invoice>Content for ` + inv.series + `</Invoice>`)
		
		// Sign document
		signedXML, err := client.SignXML(xmlContent)
		if err != nil {
			result := fmt.Sprintf("❌ %s: Sign failed - %v", inv.series, err)
			results = append(results, result)
			fmt.Println(result)
			continue
		}
		
		// User decision: send immediately or queue for later
		sendNow := true // This could be user input or business logic
		
		if sendNow {
			response, err := client.SendToSUNAT(signedXML, inv.docType, inv.series)
			if err != nil {
				result := fmt.Sprintf("⚠️ %s: Send failed - %v", inv.series, err)
				results = append(results, result)
				fmt.Println(result)
				continue
			}
			
			if response.Success {
				result := fmt.Sprintf("✅ %s: Accepted by SUNAT", inv.series)
				results = append(results, result)
				fmt.Println(result)
			} else {
				result := fmt.Sprintf("❌ %s: Rejected - %s", inv.series, response.Message)
				results = append(results, result)
				fmt.Println(result)
			}
		} else {
			// Save for later transmission
			filename := fmt.Sprintf("signed_%s.xml", inv.series)
			os.WriteFile(filename, signedXML, 0644)
			result := fmt.Sprintf("💾 %s: Signed and saved for later", inv.series)
			results = append(results, result)
			fmt.Println(result)
		}
	}
	
	fmt.Println("\n📊 Batch processing summary:")
	for _, result := range results {
		fmt.Println("  " + result)
	}
}

// Example 3: Sign now, send later pattern
func signNowSendLaterExample() {
	fmt.Println("\n3️⃣ Sign now, send later example:")
	
	client := sunatlib.NewSUNATClient(
		"20123456789",
		"MODDATOS",
		"moddatos", 
		"https://e-beta.sunat.gob.pe/ol-ti-itcpfegem-beta/billService",
	)
	defer client.Cleanup()
	
	// Phase 1: Sign documents (perhaps during business hours)
	fmt.Println("📅 Phase 1: Signing documents for later transmission...")
	
	pendingDocuments := []struct {
		content string
		docType string
		series  string
	}{
		{"<Invoice>Content 1</Invoice>", "01", "F001-00000010"},
		{"<Invoice>Content 2</Invoice>", "01", "F001-00000011"},
	}
	
	signedQueue := make([]struct {
		signedXML []byte
		docType   string
		series    string
	}, 0)
	
	for _, doc := range pendingDocuments {
		fmt.Printf("🔏 Signing %s...\n", doc.series)
		
		signedXML, err := client.SignXML([]byte(doc.content))
		if err != nil {
			fmt.Printf("❌ Failed to sign %s: %v\n", doc.series, err)
			continue
		}
		
		signedQueue = append(signedQueue, struct {
			signedXML []byte
			docType   string
			series    string
		}{signedXML, doc.docType, doc.series})
		
		fmt.Printf("✅ %s signed and queued\n", doc.series)
	}
	
	// Phase 2: Send when ready (perhaps during different time window)
	fmt.Println("\n📅 Phase 2: Sending queued documents to SUNAT...")
	time.Sleep(1 * time.Second) // Simulate time passing
	
	for _, queued := range signedQueue {
		fmt.Printf("🚀 Sending %s to SUNAT...\n", queued.series)
		
		response, err := client.SendToSUNAT(queued.signedXML, queued.docType, queued.series)
		if err != nil {
			fmt.Printf("❌ Failed to send %s: %v\n", queued.series, err)
			continue
		}
		
		if response.Success {
			fmt.Printf("✅ %s accepted by SUNAT\n", queued.series)
		} else {
			fmt.Printf("❌ %s rejected: %s\n", queued.series, response.Message)
		}
	}
	
	fmt.Println("\n🎯 Benefits of this approach:")
	fmt.Println("  • Sign documents when certificates are available")
	fmt.Println("  • Send during optimal network conditions")
	fmt.Println("  • Handle offline/online scenarios")
	fmt.Println("  • Implement retry logic for failed transmissions")
	fmt.Println("  • Separate concerns: signing vs transmission")
}