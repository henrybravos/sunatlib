package gre

import (
	"os"
	"strings"
	"testing"
)

func TestGenerateExampleXMLs(t *testing.T) {
	// 1. Private Transport Example
	privateGuide := &DespatchAdvice{
		ID:        "T001-00000001",
		IssueDate: "2023-10-27",
		IssueTime: "12:00:00",
		TypeCode:  "09",
		Signature: Signature{ID: "IDSignKG", SignatoryParty: Party{PartyIdentification: ID{ID: "20600000000"}, PartyName: Name{Name: "MI EMPRESA SAC"}}},
		DespatchSupplierParty: SupplierParty{CustomerAssignedAccountID: "20600000000", AdditionalAccountID: "6", Party: Party{PartyIdentification: ID{ID: "20600000000"}, PartyName: Name{Name: "MI EMPRESA SAC"}}},
		DeliveryCustomerParty: CustomerParty{CustomerAssignedAccountID: "20400000000", AdditionalAccountID: "6", Party: Party{PartyIdentification: ID{ID: "20400000000"}, PartyName: Name{Name: "CLIENTE SAC"}}},
		Shipment: Shipment{
			HandlingCode: "02",
			GrossWeightMeasure: Measure{Value: 150.5, UnitCode: "KGM"},
			ShipmentStages: []ShipmentStage{
				{
					ID:                "1",
					TransportModeCode: "02",
					TransitPeriod:     Period{StartDate: "2023-10-28"},
					TransportMeans:    &TransportMeans{RoadTransportInstallation: RoadInstallation{LicensePlateID: "ABC-123"}},
					DriverPerson:      &Person{ID: ID{ID: "10203040"}},
				},
			},
			Delivery: Delivery{
				Despatch: Despatch{DespatchAddress: Address{ID: "150101", AddressLine: Line{Line: "AV. LOS PINOS 123"}}},
				DeliveryAddress: Address{ID: "150142", AddressLine: Line{Line: "CALLE LAS FLORES 456"}},
			},
		},
		DespatchLines: []DespatchLine{
			{
				ID:                "1",
				DeliveredQuantity: Measure{Value: 10, UnitCode: "NIU"},
				Item:              Item{Description: "PRODUCTO DE PRUEBA 1", ID: ID{ID: "PROD001"}},
			},
		},
	}

	// 2. Public Transport Example
	publicGuide := &DespatchAdvice{
		ID:        "T001-00000002",
		IssueDate: "2023-10-27",
		IssueTime: "12:05:00",
		TypeCode:  "09",
		Signature: Signature{ID: "IDSignKG", SignatoryParty: Party{PartyIdentification: ID{ID: "20600000000"}, PartyName: Name{Name: "MI EMPRESA SAC"}}},
		DespatchSupplierParty: SupplierParty{CustomerAssignedAccountID: "20600000000", AdditionalAccountID: "6", Party: Party{PartyIdentification: ID{ID: "20600000000"}, PartyName: Name{Name: "MI EMPRESA SAC"}}},
		DeliveryCustomerParty: CustomerParty{CustomerAssignedAccountID: "20400000000", AdditionalAccountID: "6", Party: Party{PartyIdentification: ID{ID: "20400000000"}, PartyName: Name{Name: "CLIENTE SAC"}}},
		Shipment: Shipment{
			HandlingCode: "01",
			GrossWeightMeasure: Measure{Value: 200.0, UnitCode: "KGM"},
			ShipmentStages: []ShipmentStage{
				{
					ID:                "1",
					TransportModeCode: "01",
					TransitPeriod:     Period{StartDate: "2023-10-28"},
					CarrierParty:      &CarrierParty{PartyIdentification: ID{ID: "20601234567"}, PartyName: Name{Name: "TRANSPORTES EXPRESS SAC"}},
				},
			},
			Delivery: Delivery{
				Despatch: Despatch{DespatchAddress: Address{ID: "150101", AddressLine: Line{Line: "AV. LOS PINOS 123"}}},
				DeliveryAddress: Address{ID: "150142", AddressLine: Line{Line: "CALLE LAS FLORES 456"}},
			},
		},
		DespatchLines: []DespatchLine{
			{
				ID:                "1",
				DeliveredQuantity: Measure{Value: 5, UnitCode: "NIU"},
				Item:              Item{Description: "PRODUCTO DE PRUEBA 2", ID: ID{ID: "PROD002"}},
			},
		},
	}

	// 3. Establishment Transfer Example
	transferGuide := &DespatchAdvice{
		ID:        "T001-00000003",
		IssueDate: "2023-10-27",
		IssueTime: "12:10:00",
		TypeCode:  "09",
		Signature: Signature{ID: "IDSignKG", SignatoryParty: Party{PartyIdentification: ID{ID: "20600000000"}, PartyName: Name{Name: "MI EMPRESA SAC"}}},
		DespatchSupplierParty: SupplierParty{CustomerAssignedAccountID: "20600000000", AdditionalAccountID: "6", Party: Party{PartyIdentification: ID{ID: "20600000000"}, PartyName: Name{Name: "MI EMPRESA SAC"}}},
		DeliveryCustomerParty: CustomerParty{CustomerAssignedAccountID: "20600000000", AdditionalAccountID: "6", Party: Party{PartyIdentification: ID{ID: "20600000000"}, PartyName: Name{Name: "MI EMPRESA SAC"}}},
		Shipment: Shipment{
			HandlingCode: "04", // Traslado entre establecimientos de la misma empresa
			GrossWeightMeasure: Measure{Value: 50.0, UnitCode: "KGM"},
			ShipmentStages: []ShipmentStage{
				{
					ID:                "1",
					TransportModeCode: "02",
					TransitPeriod:     Period{StartDate: "2023-10-28"},
					TransportMeans:    &TransportMeans{RoadTransportInstallation: RoadInstallation{LicensePlateID: "ABC-123"}},
					DriverPerson:      &Person{ID: ID{ID: "10203040"}},
				},
			},
			Delivery: Delivery{
				Despatch: Despatch{DespatchAddress: Address{ID: "150101", AddressLine: Line{Line: "ALMACEN CENTRAL"}}},
				DeliveryAddress: Address{ID: "150142", AddressLine: Line{Line: "SUCURSAL NORTE"}},
			},
		},
		DespatchLines: []DespatchLine{
			{
				ID:                "1",
				DeliveredQuantity: Measure{Value: 20, UnitCode: "NIU"},
				Item:              Item{Description: "STOCK TRANSFER", ID: ID{ID: "STOCK001"}},
			},
		},
	}

	examples := []struct {
		name  string
		guide *DespatchAdvice
		file  string
	}{
		{"Private Example", privateGuide, "../testdata/gre/examples/guide_private.xml"},
		{"Public Example", publicGuide, "../testdata/gre/examples/guide_public.xml"},
		{"Transfer Example", transferGuide, "../testdata/gre/examples/guide_transfer.xml"},
	}

	for _, ex := range examples {
		xml, err := GenerateXML(ex.guide)
		if err != nil {
			t.Fatalf("Failed to generate %s: %v", ex.name, err)
		}
		// In a real environment we would write to file, here we just verify it doesn't error
		// and we can manually check the output if needed.
		t.Logf("Generated %s successfully", ex.name)
		
		// To actually save them during test:
		err = os.WriteFile(ex.file, xml, 0644)
		if err != nil {
			t.Errorf("Failed to save %s: %v", ex.name, err)
		}
	}
}

func contains(s, substr string) bool {
	return strings.Index(s, substr) >= 0
}

func TestGenerateXML(t *testing.T) {
	tests := []struct {
		name    string
		guide   *DespatchAdvice
		wantErr bool
		checks  []string
	}{
		{
			name: "Private Transport (Remitente)",
			guide: &DespatchAdvice{
				ID: "T001-1",
				Shipment: Shipment{
					HandlingCode: "02", // Privado
					ShipmentStages: []ShipmentStage{
						{
							TransportMeans: &TransportMeans{
								RoadTransportInstallation: RoadInstallation{
									LicensePlateID: "ABC-123",
								},
							},
							DriverPerson: &Person{
								ID: ID{ID: "10203040"},
							},
						},
					},
				},
			},
			wantErr: false,
			checks: []string{
				"<cbc:ID>T001-1</cbc:ID>",
				"<cbc:HandlingCode>02</cbc:HandlingCode>",
				"<cbc:LicensePlateID>ABC-123</cbc:LicensePlateID>",
				"<cac:DriverPerson>",
				"<cbc:ID>10203040</cbc:ID>",
			},
		},
		{
			name: "Public Transport (Transportista)",
			guide: &DespatchAdvice{
				ID: "T001-2",
				Shipment: Shipment{
					HandlingCode: "01", // Público
					ShipmentStages: []ShipmentStage{
						{
							CarrierParty: &CarrierParty{
								PartyIdentification: ID{ID: "20600000001"},
								PartyName:           Name{Name: "Transportes SAC"},
							},
						},
					},
				},
			},
			wantErr: false,
			checks: []string{
				"<cbc:ID>T001-2</cbc:ID>",
				"<cbc:HandlingCode>01</cbc:HandlingCode>",
				"<cac:CarrierParty>",
				"<cbc:ID>20600000001</cbc:ID>",
				"<cbc:Name><![CDATA[Transportes SAC]]></cbc:Name>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xml, err := GenerateXML(tt.guide)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateXML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				t.Logf("Generated XML for %s:\n%s", tt.name, string(xml))
				for _, check := range tt.checks {
					if !strings.Contains(string(xml), check) {
						t.Errorf("GenerateXML() missing expected string: %s", check)
					}
				}
			}
		})
	}
}
