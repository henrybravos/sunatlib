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
        <cbc:CustomerAssignedAccountID>%s</cbc:CustomerAssignedAccountID>
        <cbc:AdditionalAccountID>%s</cbc:AdditionalAccountID>
        <cac:Party>
            <cac:PartyIdentification>
                <cbc:ID>%s</cbc:ID>
            </cac:PartyIdentification>
            <cac:PartyName>
                <cbc:Name><![CDATA[%s]]></cbc:Name>
            </cac:PartyName>
        </cac:Party>
    </cac:DespatchSupplierParty>
    <cac:DeliveryCustomerParty>
        <cbc:CustomerAssignedAccountID>%s</cbc:CustomerAssignedAccountID>
        <cbc:AdditionalAccountID>%s</cbc:AdditionalAccountID>
        <cac:Party>
            <cac:PartyIdentification>
                <cbc:ID>%s</cbc:ID>
            </cac:PartyIdentification>
            <cac:PartyName>
                <cbc:Name><![CDATA[%s]]></cbc:Name>
            </cac:PartyName>
        </cac:Party>
    </cac:DeliveryCustomerParty>
    <cac:Shipment>
        <cbc:ID>1</cbc:ID>
        <cbc:HandlingCode>%s</cbc:HandlingCode>
        <cbc:GrossWeightMeasure unitCode="%s">%f</cbc:GrossWeightMeasure>
        %s
        <cac:Delivery>
            <cac:Despatch>
                <cac:DespatchAddress>
                    <cbc:ID>%s</cbc:ID>
                    <cac:AddressLine>
                        <cbc:Line><![CDATA[%s]]></cbc:Line>
                    </cac:AddressLine>
                </cac:DespatchAddress>
            </cac:Despatch>
            <cac:DeliveryAddress>
                <cbc:ID>%s</cbc:ID>
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
			carrierXML = fmt.Sprintf(`
        <cac:CarrierParty>
            <cac:PartyIdentification>
                <cbc:ID>%s</cbc:ID>
            </cac:PartyIdentification>
            <cac:PartyName>
                <cbc:Name><![CDATA[%s]]></cbc:Name>
            </cac:PartyName>
        </cac:CarrierParty>`,
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
			driverXML = fmt.Sprintf(`
        <cac:DriverPerson>
            <cac:ID>
                <cbc:ID>%s</cbc:ID>
            </cac:ID>
        </cac:DriverPerson>`,
				stage.DriverPerson.ID.ID,
			)
		}

		stagesXML += fmt.Sprintf(`
    <cac:ShipmentStage>
        <cbc:ID>%s</cbc:ID>
        <cbc:TransportModeCode>%s</cbc:TransportModeCode>
        <cac:TransitPeriod>
            <cbc:StartDate>%s</cbc:StartDate>
        </cac:TransitPeriod>%s%s%s
    </cac:ShipmentStage>`,
			stage.ID,
			stage.TransportModeCode,
			stage.TransitPeriod.StartDate,
			carrierXML,
			meansXML,
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

	xmlContent := fmt.Sprintf(despatchAdviceTemplate,
		guide.ID,
		guide.IssueDate,
		guide.IssueTime,
		guide.TypeCode,
		guide.Signature.ID,
		guide.Signature.SignatoryParty.PartyIdentification.ID,
		guide.Signature.SignatoryParty.PartyName.Name,
		guide.Signature.ID, // URI #ID
		guide.DespatchSupplierParty.CustomerAssignedAccountID,
		guide.DespatchSupplierParty.AdditionalAccountID,
		guide.DespatchSupplierParty.Party.PartyIdentification.ID,
		guide.DespatchSupplierParty.Party.PartyName.Name,
		guide.DeliveryCustomerParty.CustomerAssignedAccountID,
		guide.DeliveryCustomerParty.AdditionalAccountID,
		guide.DeliveryCustomerParty.Party.PartyIdentification.ID,
		guide.DeliveryCustomerParty.Party.PartyName.Name,
		guide.Shipment.HandlingCode,
		guide.Shipment.GrossWeightMeasure.UnitCode,
		guide.Shipment.GrossWeightMeasure.Value,
		stagesXML,
		guide.Shipment.Delivery.Despatch.DespatchAddress.ID,
		guide.Shipment.Delivery.Despatch.DespatchAddress.AddressLine.Line,
		guide.Shipment.Delivery.DeliveryAddress.ID,
		guide.Shipment.Delivery.DeliveryAddress.AddressLine.Line,
		linesXML,
	)

	return []byte(xmlContent), nil
}
