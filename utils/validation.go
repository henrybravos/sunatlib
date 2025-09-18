// Package utils provides validation utilities for SUNAT documents
package utils

import (
	"html"
	"regexp"
	"strings"
)

// ValidateSpecialCharacters validates and cleans special characters for SUNAT XML
// This is equivalent to the ValidarCaracteresInv function in PHP
func ValidateSpecialCharacters(text string) string {
	if text == "" {
		return text
	}

	// Remove or replace problematic characters
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	text = strings.ReplaceAll(text, "\"", "&quot;")
	text = strings.ReplaceAll(text, "'", "&apos;")

	// Remove control characters and non-printable characters
	re := regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
	text = re.ReplaceAllString(text, "")

	// Trim spaces
	text = strings.TrimSpace(text)

	return text
}

// CleanTextForXML prepares text for safe inclusion in XML
func CleanTextForXML(text string) string {
	// First validate special characters
	text = ValidateSpecialCharacters(text)

	// HTML escape to ensure XML safety
	text = html.EscapeString(text)

	return text
}

// ValidateRUC validates a Peruvian RUC number format
func ValidateRUC(ruc string) bool {
	if len(ruc) != 11 {
		return false
	}

	// Check if all characters are digits
	re := regexp.MustCompile(`^\d{11}$`)
	if !re.MatchString(ruc) {
		return false
	}

	// Validate RUC checksum (basic validation)
	factors := []int{5, 4, 3, 2, 7, 6, 5, 4, 3, 2}
	sum := 0

	for i, digit := range ruc[:10] {
		digitValue := int(digit - '0')
		sum += digitValue * factors[i]
	}

	remainder := sum % 11
	checkDigit := 11 - remainder

	if checkDigit == 10 {
		checkDigit = 0
	} else if checkDigit == 11 {
		checkDigit = 1
	}

	expectedCheckDigit := int(ruc[10] - '0')
	return checkDigit == expectedCheckDigit
}

// ValidateDocumentSeries validates a document series format
func ValidateDocumentSeries(series string) bool {
	if len(series) < 3 || len(series) > 4 {
		return false
	}

	// Common series patterns: F001, B001, FC01, etc.
	re := regexp.MustCompile(`^[A-Z]{1,2}\d{2,3}$`)
	return re.MatchString(series)
}

// ValidateDocumentNumber validates a document number format
func ValidateDocumentNumber(number string) bool {
	if len(number) == 0 || len(number) > 8 {
		return false
	}

	// Should be numeric
	re := regexp.MustCompile(`^\d+$`)
	return re.MatchString(number)
}

// ValidateDocumentType validates document type codes
func ValidateDocumentType(docType string) bool {
	validTypes := map[string]bool{
		"01": true, // Factura
		"03": true, // Boleta de Venta
		"07": true, // Nota de Crédito
		"08": true, // Nota de Débito
		"09": true, // Guía de Remisión Remitente
		"31": true, // Guía de Remisión Transportista
		"04": true, // Liquidación de Compra
	}

	return validTypes[docType]
}

// GenerateLineID generates a line ID for voided documents
func GenerateLineID(index int) int {
	return index + 1
}