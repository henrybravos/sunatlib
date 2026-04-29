package sunatlib

import (
	"strings"
	"testing"
)

func TestUBLValidator_Validate(t *testing.T) {
	v := NewUBLValidator()

	tests := []struct {
		name    string
		xml     string
		wantErr bool
		msg     string
	}{
		{
			name: "Valid Invoice",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2">
	<cbc:UBLVersionID>2.1</cbc:UBLVersionID>
	<cbc:ID>F001-1</cbc:ID>
	<cac:InvoiceLine>
		<cbc:ID>1</cbc:ID>
		<cac:TaxTotal>
			<cac:TaxSubtotal>
				<cac:TaxCategory>
					<cac:TaxScheme><cbc:ID>1000</cbc:ID></cac:TaxScheme>
				</cac:TaxCategory>
			</cac:TaxSubtotal>
		</cac:TaxTotal>
	</cac:InvoiceLine>
</Invoice>`,
			wantErr: false,
		},
		{
			name: "Missing TaxTotal",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2">
	<cbc:UBLVersionID>2.1</cbc:UBLVersionID>
	<cbc:ID>F001-1</cbc:ID>
	<cac:InvoiceLine>
		<cbc:ID>1</cbc:ID>
	</cac:InvoiceLine>
</Invoice>`,
			wantErr: true,
			msg:     "missing <cac:TaxTotal>",
		},
		{
			name: "Invalid TaxScheme",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2">
	<cbc:UBLVersionID>2.1</cbc:UBLVersionID>
	<cbc:ID>F001-1</cbc:ID>
	<cac:InvoiceLine>
		<cbc:ID>1</cbc:ID>
		<cac:TaxTotal>
			<cac:TaxSubtotal>
				<cac:TaxCategory>
					<cac:TaxScheme><cbc:ID>999</cbc:ID></cac:TaxScheme>
				</cac:TaxCategory>
			</cac:TaxSubtotal>
		</cac:TaxTotal>
	</cac:InvoiceLine>
</Invoice>`,
			wantErr: true,
			msg:     "invalid TaxScheme ID",
		},
		{
			name: "Gratuity (Zero Amount but Valid)",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2">
	<cbc:UBLVersionID>2.1</cbc:UBLVersionID>
	<cbc:ID>F001-1</cbc:ID>
	<cac:InvoiceLine>
		<cbc:ID>1</cbc:ID>
		<cac:TaxTotal>
			<cbc:TaxAmount>0.00</cbc:TaxAmount>
			<cac:TaxSubtotal>
				<cbc:TaxableAmount>0.00</cbc:TaxableAmount>
				<cbc:TaxAmount>0.00</cbc:TaxAmount>
				<cac:TaxCategory>
					<cbc:TaxExemptionReasonCode>31</cbc:TaxExemptionReasonCode>
					<cac:TaxScheme><cbc:ID>1000</cbc:ID></cac:TaxScheme>
				</cac:TaxCategory>
			</cac:TaxSubtotal>
		</cac:TaxTotal>
	</cac:InvoiceLine>
</Invoice>`,
			wantErr: false,
		},
		{
			name: "Valid DespatchAdvice",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<DespatchAdvice xmlns="urn:oasis:names:specification:ubl:schema:xsd:DespatchAdvice-2">
	<cbc:UBLVersionID>2.1</cbc:UBLVersionID>
	<cbc:CustomizationID>2.0</cbc:CustomizationID>
	<cbc:ID>T001-1</cbc:ID>
	<cac:DespatchLine>
		<cbc:ID>1</cbc:ID>
		<cbc:DeliveredQuantity unitCode="NIU">10</cbc:DeliveredQuantity>
		<cac:Item>
			<cbc:Description>Item Test</cbc:Description>
		</cac:Item>
	</cac:DespatchLine>
</DespatchAdvice>`,
			wantErr: false,
		},
		{
			name: "DespatchAdvice Invalid UBLVersion",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<DespatchAdvice xmlns="urn:oasis:names:specification:ubl:schema:xsd:DespatchAdvice-2">
	<cbc:UBLVersionID>2.0</cbc:UBLVersionID>
	<cbc:CustomizationID>2.0</cbc:CustomizationID>
	<cbc:ID>T001-1</cbc:ID>
</DespatchAdvice>`,
			wantErr: true,
			msg:     "UBL error: GRE must use UBLVersionID 2.1",
		},
		{
			name: "DespatchAdvice Invalid CustomizationID",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<DespatchAdvice xmlns="urn:oasis:names:specification:ubl:schema:xsd:DespatchAdvice-2">
	<cbc:UBLVersionID>2.1</cbc:UBLVersionID>
	<cbc:CustomizationID>1.0</cbc:CustomizationID>
	<cbc:ID>T001-1</cbc:ID>
</DespatchAdvice>`,
			wantErr: true,
			msg:     "UBL error: GRE must use CustomizationID 2.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate([]byte(tt.xml))
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.msg != "" {
				if !contains(err.Error(), tt.msg) {
					t.Errorf("Validate() error = %v, want msg containing %v", err, tt.msg)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return strings.Index(s, substr) >= 0
}
