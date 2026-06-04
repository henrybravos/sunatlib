package gre

import (
	"fmt"
)

const despatchAdviceTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<DespatchAdvice xmlns="urn:oasis:names:specification:ubl:schema:xsd:DespatchAdvice-2" 
                xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2" 
                xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2" 
                xmlns:ds="http://www.w3.org/2000/09/xmldsig#" 
                xmlns:ext="urn:oasis:names:specification:ubl:schema:xsd:CommonExtensionComponents-2">
    <ext:UBLExtensions>
        <ext:UBLExtension>
            <ext:ExtensionContent/>
        </ext:UBLExtension>
    </ext:UBLExtensions>
    <cbc:UBLVersionID>2.1</cbc:UBLVersionID>
    <cbc:CustomizationID>2.0</cbc:CustomizationID>
    <cbc:ID>%s</cbc:ID>
    <cbc:IssueDate>%s</cbc:IssueDate>
    <cbc:IssueTime>%s</cbc:IssueTime>
    <cbc:DespatchAdviceTypeCode>%s</cbc:DespatchAdviceTypeCode>
    <cac:Signature>
        <cbc:ID>%s</cbc:ID>
        <cac:SignatoryParty>
            <cac:PartyIdentification>
                <cbc:ID>%s</cbc:ID>
            </cac:PartyIdentification>
            <cac:PartyName>
                <cbc:Name><![CDATA[%s]]></cbc:Name>
            </cac:PartyName>
        </cac:SignatoryParty>
        <cac:DigitalSignatureAttachment>
            <cac:ExternalReference>
                <cbc:URI>#%s</cbc:URI>
            </cac:ExternalReference>
        </cac:DigitalSignatureAttachment>
    </cac:Signature>
    <cac:DespatchSupplierParty>
        <cac:Party>
            <cac:PartyIdentification>
                <cbc:ID schemeID="6">%s</cbc:ID>
            </cac:PartyIdentification>
            <cac:PartyLegalEntity>
                <cbc:RegistrationName><![CDATA[%s]]></cbc:RegistrationName>
            </cac:PartyLegalEntity>
        </cac:Party>
    </cac:DespatchSupplierParty>
    <cac:DeliveryCustomerParty>
        <cac:Party>
            <cac:PartyIdentification>
                <cbc:ID schemeID="%s">%s</cbc:ID>
            </cac:PartyIdentification>
            <cac:PartyLegalEntity>
                <cbc:RegistrationName><![CDATA[%s]]></cbc:RegistrationName>
            </cac:PartyLegalEntity>
        </cac:Party>
    </cac:DeliveryCustomerParty>
    <cac:Shipment>
        <cbc:ID>%s</cbc:ID>
        <cbc:HandlingCode>%s</cbc:HandlingCode>
        %s
        <cbc:GrossWeightMeasure unitCode="%s">%f</cbc:GrossWeightMeasure>
        %s
        <cac:Delivery>
            <cac:Despatch>
                <cac:DespatchAddress>
                    <cbc:ID>%s</cbc:ID>
                    %s
                    <cac:AddressLine>
                        <cbc:Line><![CDATA[%s]]></cbc:Line>
                    </cac:AddressLine>
                </cac:DespatchAddress>
            </cac:Despatch>
            <cac:DeliveryAddress>
                <cbc:ID>%s</cbc:ID>
                %s
                <cac:AddressLine>
                    <cbc:Line><![CDATA[%s]]></cbc:Line>
                </cac:AddressLine>
            </cac:DeliveryAddress>
        </cac:Delivery>
    </cac:Shipment>
    %s
</DespatchAdvice>`

// GenerateXML generates the UBL 2.1 DespatchAdvice XML using a template for precision
func GenerateXML(guide *DespatchAdvice) ([]byte, error) {
	stagesXML := ""
	for _, stage := range guide.Shipment.ShipmentStages {
		carrierXML := ""
		if stage.CarrierParty != nil {
			carrierDocType := "6"
			if len(stage.CarrierParty.PartyIdentification.ID) != 11 {
				carrierDocType = "1"
			}
			carrierXML = fmt.Sprintf(`
        <cac:CarrierParty>
            <cac:PartyIdentification>
                <cbc:ID schemeID="%s">%s</cbc:ID>
            </cac:PartyIdentification>
            <cac:PartyLegalEntity>
                <cbc:RegistrationName><![CDATA[%s]]></cbc:RegistrationName>
            </cac:PartyLegalEntity>
        </cac:CarrierParty>`,
				carrierDocType,
				stage.CarrierParty.PartyIdentification.ID,
				stage.CarrierParty.PartyName.Name,
			)
		}

		meansXML := ""
		if stage.TransportMeans != nil {
			meansXML = fmt.Sprintf(`
        <cac:TransportMeans>
            <cac:RoadTransportInstallation>
                <cbc:LicensePlateID>%s</cbc:LicensePlateID>
            </cac:RoadTransportInstallation>
        </cac:TransportMeans>`,
				stage.TransportMeans.RoadTransportInstallation.LicensePlateID,
			)
		}

		driverXML := ""
		if stage.DriverPerson != nil {
			driverDocType := "1"
			if len(stage.DriverPerson.ID.ID) == 11 {
				driverDocType = "6"
			} else if len(stage.DriverPerson.ID.ID) != 8 {
				driverDocType = "4"
			}
			driverXML = fmt.Sprintf(`
        <cac:DriverPerson>
            <cbc:ID schemeID="%s">%s</cbc:ID>
        </cac:DriverPerson>`,
				driverDocType,
				stage.DriverPerson.ID.ID,
			)
		}

		loadingXML := ""
		if stage.LoadingTransportEvent != nil && stage.LoadingTransportEvent.OccurrenceDate != "" {
			loadingXML = fmt.Sprintf(`
        <cac:LoadingTransportEvent>
            <cbc:OccurrenceDate>%s</cbc:OccurrenceDate>
        </cac:LoadingTransportEvent>`,
				stage.LoadingTransportEvent.OccurrenceDate,
			)
		}

		stagesXML += fmt.Sprintf(`
    <cac:ShipmentStage>
        <cbc:ID>%s</cbc:ID>
        <cbc:TransportModeCode>%s</cbc:TransportModeCode>
        <cac:TransitPeriod>
            <cbc:StartDate>%s</cbc:StartDate>
        </cac:TransitPeriod>%s%s%s%s
    </cac:ShipmentStage>`,
			stage.ID,
			stage.TransportModeCode,
			stage.TransitPeriod.StartDate,
			carrierXML,
			meansXML,
			loadingXML,
			driverXML,
		)
	}

	linesXML := ""
	for _, line := range guide.DespatchLines {
		linesXML += fmt.Sprintf(`
    <cac:DespatchLine>
        <cbc:ID>%s</cbc:ID>
        <cbc:DeliveredQuantity unitCode="%s">%f</cbc:DeliveredQuantity>
        <cac:Item>
            <cbc:Description><![CDATA[%s]]></cbc:Description>
            <cac:SellersItemIdentification>
                <cbc:ID>%s</cbc:ID>
            </cac:SellersItemIdentification>
        </cac:Item>
    </cac:DespatchLine>`,
			line.ID,
			line.DeliveredQuantity.UnitCode,
			line.DeliveredQuantity.Value,
			line.Item.Description,
			line.Item.ID.ID,
		)
	}

	customerDocType := "6"
	if len(guide.DeliveryCustomerParty.Party.PartyIdentification.ID) != 11 {
		customerDocType = "1"
	}

	specialInstructionsXML := ""
	if guide.Shipment.SpecialInstructions != "" {
		specialInstructionsXML = fmt.Sprintf("\n        <cbc:SpecialInstructions>%s</cbc:SpecialInstructions>", guide.Shipment.SpecialInstructions)
	}

	shipmentID := guide.Shipment.ID
	if shipmentID == "" {
		shipmentID = "1"
	}

	despatchAddressTypeCodeXML := ""
	if guide.Shipment.Delivery.Despatch.DespatchAddress.AddressTypeCode != "" {
		despatchAddressTypeCodeXML = fmt.Sprintf("\n                    <cbc:AddressTypeCode>%s</cbc:AddressTypeCode>", guide.Shipment.Delivery.Despatch.DespatchAddress.AddressTypeCode)
	}

	deliveryAddressTypeCodeXML := ""
	if guide.Shipment.Delivery.DeliveryAddress.AddressTypeCode != "" {
		deliveryAddressTypeCodeXML = fmt.Sprintf("\n                <cbc:AddressTypeCode>%s</cbc:AddressTypeCode>", guide.Shipment.Delivery.DeliveryAddress.AddressTypeCode)
	}

	xmlContent := fmt.Sprintf(despatchAdviceTemplate,
		guide.ID,
		guide.IssueDate,
		guide.IssueTime,
		guide.TypeCode,
		guide.Signature.ID,
		guide.Signature.SignatoryParty.PartyIdentification.ID,
		guide.Signature.SignatoryParty.PartyName.Name,
		guide.Signature.ID, // URI #ID
		guide.DespatchSupplierParty.Party.PartyIdentification.ID,
		guide.DespatchSupplierParty.Party.PartyName.Name,
		customerDocType,
		guide.DeliveryCustomerParty.Party.PartyIdentification.ID,
		guide.DeliveryCustomerParty.Party.PartyName.Name,
		shipmentID,
		guide.Shipment.HandlingCode,
		specialInstructionsXML,
		guide.Shipment.GrossWeightMeasure.UnitCode,
		guide.Shipment.GrossWeightMeasure.Value,
		stagesXML,
		guide.Shipment.Delivery.Despatch.DespatchAddress.ID,
		despatchAddressTypeCodeXML,
		guide.Shipment.Delivery.Despatch.DespatchAddress.AddressLine.Line,
		guide.Shipment.Delivery.DeliveryAddress.ID,
		deliveryAddressTypeCodeXML,
		guide.Shipment.Delivery.DeliveryAddress.AddressLine.Line,
		linesXML,
	)

	return []byte(xmlContent), nil
}
