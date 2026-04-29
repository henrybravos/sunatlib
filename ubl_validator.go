// Package sunatlib provides structural UBL validation for SUNAT documents
package sunatlib

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// UBLValidator handles structural validation of UBL XML documents
type UBLValidator struct{}

// NewUBLValidator creates a new UBL structural validator
func NewUBLValidator() *UBLValidator {
	return &UBLValidator{}
}

// MinimalUBL represents a minimal structure of a UBL document for validation
type MinimalUBL struct {
	XMLName         xml.Name
	UBLVersionID    string        `xml:"UBLVersionID"`
	CustomizationID string        `xml:"CustomizationID"`
	ID              string        `xml:"ID"`
	IssueDate       string        `xml:"IssueDate"`
	InvoiceLines    []InvoiceLine `xml:"InvoiceLine"`
	CreditNoteLines []InvoiceLine `xml:"CreditNoteLine"`
	DebitNoteLines  []InvoiceLine `xml:"DebitNoteLine"`
	DespatchLines   []DespatchLine `xml:"DespatchLine"`
}

// DespatchLine represents a line in a Referral Guide
type DespatchLine struct {
	ID                string  `xml:"ID"`
	DeliveredQuantity float64 `xml:"DeliveredQuantity"`
}

// InvoiceLine represents a line in a UBL document (shared between Invoice, CN, DN)
type InvoiceLine struct {
	ID               string           `xml:"ID"`
	InvoicedQuantity float64          `xml:"InvoicedQuantity"`
	LineExtensionAmount float64       `xml:"LineExtensionAmount"`
	TaxTotals        []TaxTotal       `xml:"TaxTotal"`
	Item             Item             `xml:"Item"`
	Price            Price            `xml:"Price"`
}

// TaxTotal represents a tax total block
type TaxTotal struct {
	TaxAmount    float64     `xml:"TaxAmount"`
	TaxSubtotals []TaxSubtotal `xml:"TaxSubtotal"`
}

// TaxSubtotal represents a tax subtotal block
type TaxSubtotal struct {
	TaxableAmount float64     `xml:"TaxableAmount"`
	TaxAmount     float64     `xml:"TaxAmount"`
	TaxCategory   TaxCategory `xml:"TaxCategory"`
}

// TaxCategory represents a tax category block
type TaxCategory struct {
	ID                     string    `xml:"ID"`
	Percent                float64   `xml:"Percent"`
	TaxExemptionReasonCode string    `xml:"TaxExemptionReasonCode"`
	TaxScheme              TaxScheme `xml:"TaxScheme"`
}

// TaxScheme represents a tax scheme block
type TaxScheme struct {
	ID   string `xml:"ID"`
	Name string `xml:"Name"`
}

// Item represents an item in a line
type Item struct {
	Description string `xml:"Description"`
}

// Price represents the price of an item
type Price struct {
	PriceAmount float64 `xml:"PriceAmount"`
}

// Validate performs structural validation on UBL XML content
func (v *UBLValidator) Validate(xmlContent []byte) error {
	var ubl MinimalUBL
	if err := xml.Unmarshal(xmlContent, &ubl); err != nil {
		return fmt.Errorf("failed to parse XML: %w", err)
	}

	// 1. Validate Header
	if err := v.validateHeader(&ubl); err != nil {
		return err
	}

	// 2. Validate Lines
	if err := v.validateLines(&ubl); err != nil {
		return err
	}

	return nil
}

func (v *UBLValidator) validateHeader(ubl *MinimalUBL) error {
	if ubl.ID == "" {
		return fmt.Errorf("UBL error: missing document ID (<cbc:ID>)")
	}

	if ubl.UBLVersionID == "" {
		return fmt.Errorf("UBL error: missing UBLVersionID")
	}

	rootName := strings.ToLower(ubl.XMLName.Local)
	if rootName == "despatchadvice" {
		if ubl.UBLVersionID != "2.1" {
			return fmt.Errorf("UBL error: GRE must use UBLVersionID 2.1")
		}
		if ubl.CustomizationID != "2.0" {
			return fmt.Errorf("UBL error: GRE must use CustomizationID 2.0")
		}
	}

	return nil
}

func (v *UBLValidator) validateLines(ubl *MinimalUBL) error {
	lines := ubl.InvoiceLines
	if len(lines) == 0 {
		lines = ubl.CreditNoteLines
	}
	if len(lines) == 0 {
		lines = ubl.DebitNoteLines
	}

	rootName := strings.ToLower(ubl.XMLName.Local)
	if rootName == "despatchadvice" {
		if len(ubl.DespatchLines) == 0 {
			return fmt.Errorf("UBL error: document must have at least one line")
		}
		return nil
	}

	if rootName != "invoice" && rootName != "creditnote" && rootName != "debitnote" {
		return nil
	}

	if len(lines) == 0 {
		return fmt.Errorf("UBL error: document must have at least one line")
	}

	for i, line := range lines {
		// Every line MUST have a TaxTotal (SUNAT Rule)
		if len(line.TaxTotals) == 0 {
			return fmt.Errorf("UBL error: line %d (ID: %s) is missing <cac:TaxTotal> (SUNAT 3105)", i+1, line.ID)
		}

		// Validate TaxScheme IDs per SUNAT Catálogo 05
		// Amounts (0 or non-zero) are NOT validated here — SUNAT handles business rules.
		// We only block structurally invalid or unknown TaxScheme codes.
		for _, taxTotal := range line.TaxTotals {
			for _, subtotal := range taxTotal.TaxSubtotals {
				schemeID := subtotal.TaxCategory.TaxScheme.ID
				if !isValidTaxScheme(schemeID) {
					return fmt.Errorf("UBL error: line %d has invalid TaxScheme ID '%s' (SUNAT Catálogo 05)", i+1, schemeID)
				}
			}
		}
	}

	return nil
}

// isValidTaxScheme checks if a TaxScheme/cbc:ID is a valid SUNAT Catálogo 05 code.
// Mapping from tipAfeIgv (Catálogo 07) → TaxScheme ID (Catálogo 05):
//
//	tipAfeIgv 10              → 1000 (IGV - Gravado Oneroso)
//	tipAfeIgv 11-16 (Retiros) → 9996 (GRA - Gravado Retiro, empresa absorbe IGV)
//	tipAfeIgv 17              → 1016 (IVAP)
//	tipAfeIgv 20,21           → 9997 (EXO - Exonerado)
//	tipAfeIgv 30-36           → 9998 (INA - Inafecto)
//	tipAfeIgv 40              → 9995 (EXP - Exportación)
//	ISC                       → 2000
//	Otros (ICBPER, etc.)      → 9999
func isValidTaxScheme(id string) bool {
	validSchemes := map[string]bool{
		"1000": true, // IGV  — Gravado Oneroso (tipAfeIgv 10)
		"1016": true, // IVAP — Gravado IVAP (tipAfeIgv 17)
		"2000": true, // ISC  — Impuesto Selectivo al Consumo
		"9995": true, // EXP  — Exportación (tipAfeIgv 40)
		"9996": true, // GRA  — Gravado Retiro/Gratuito (tipAfeIgv 11-16)
		"9997": true, // EXO  — Exonerado (tipAfeIgv 20, 21)
		"9998": true, // INA  — Inafecto (tipAfeIgv 30-36)
		"9999": true, // OTH  — Otros tributos (ICBPER, etc.)
	}
	return validSchemes[id]
}
