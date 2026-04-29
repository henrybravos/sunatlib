package gre

import (
	"encoding/xml"
	"time"
)

// DespatchAdvice represents a UBL 2.1 Referral Guide (Guía de Remisión)
type DespatchAdvice struct {
	XMLName         xml.Name `xml:"urn:oasis:names:specification:ubl:schema:xsd:DespatchAdvice-2 DespatchAdvice"`
	XmlnsCac        string   `xml:"xmlns:cac,attr"`
	XmlnsCbc        string   `xml:"xmlns:cbc,attr"`
	XmlnsDs         string   `xml:"xmlns:ds,attr"`
	XmlnsExt        string   `xml:"xmlns:ext,attr"`
	UBLVersionID    string   `xml:"cbc:UBLVersionID"`
	CustomizationID string   `xml:"cbc:CustomizationID"`
	ID              string   `xml:"cbc:ID"`
	IssueDate       string   `xml:"cbc:IssueDate"`
	IssueTime       string   `xml:"cbc:IssueTime"`
	TypeCode        string   `xml:"cbc:DespatchAdviceTypeCode"`
	Note            string   `xml:"cbc:Note,omitempty"`

	Signature Signature `xml:"cac:Signature"`

	DespatchSupplierParty SupplierParty `xml:"cac:DespatchSupplierParty"`
	DeliveryCustomerParty CustomerParty `xml:"cac:DeliveryCustomerParty"`
	Shipment              Shipment      `xml:"cac:Shipment"`
	DespatchLines         []DespatchLine `xml:"cac:DespatchLine"`
}

type Signature struct {
	ID                     string `xml:"cbc:ID"`
	SignatoryParty         Party  `xml:"cac:SignatoryParty"`
	DigitalSignatureAttach Attach `xml:"cac:DigitalSignatureAttachment"`
}

type Party struct {
	PartyIdentification ID `xml:"cac:PartyIdentification"`
	PartyName           Name `xml:"cac:PartyName"`
}

type ID struct {
	ID string `xml:"cbc:ID"`
}

type Name struct {
	Name string `xml:"cbc:Name"`
}

type Attach struct {
	ExternalReference URI `xml:"cac:ExternalReference"`
}

type URI struct {
	URI string `xml:"cbc:URI"`
}

type SupplierParty struct {
	CustomerAssignedAccountID string `xml:"cbc:CustomerAssignedAccountID"`
	AdditionalAccountID       string `xml:"cbc:AdditionalAccountID"`
	Party                     Party  `xml:"cac:Party"`
}

type CustomerParty struct {
	CustomerAssignedAccountID string `xml:"cbc:CustomerAssignedAccountID"`
	AdditionalAccountID       string `xml:"cbc:AdditionalAccountID"`
	Party                     Party  `xml:"cac:Party"`
}

type Shipment struct {
	ID                     string         `xml:"cbc:ID"`
	HandlingCode           string         `xml:"cbc:HandlingCode"` // Modalidad de Traslado: 01 Público, 02 Privado
	Information            string         `xml:"cbc:Information,omitempty"`
	SplitConsignmentIndicator bool        `xml:"cbc:SplitConsignmentIndicator"`
	GrossWeightMeasure     Measure        `xml:"cbc:GrossWeightMeasure"`
	ShipmentStages         []ShipmentStage `xml:"cac:ShipmentStage"`
	Delivery               Delivery       `xml:"cac:Delivery"`
	TransportHandlingUnit  HandlingUnit   `xml:"cac:TransportHandlingUnit"`
}

type Measure struct {
	Value    float64 `xml:",chardata"`
	UnitCode string  `xml:"unitCode,attr"`
}

type ShipmentStage struct {
	ID                      string         `xml:"cbc:ID"`
	TransportModeCode       string         `xml:"cbc:TransportModeCode"`
	TransitPeriod           Period         `xml:"cac:TransitPeriod"`
	CarrierParty            *CarrierParty  `xml:"cac:CarrierParty,omitempty"`
	TransportMeans          *TransportMeans `xml:"cac:TransportMeans,omitempty"`
	DriverPerson            *Person        `xml:"cac:DriverPerson,omitempty"`
}

type Period struct {
	StartDate string `xml:"cbc:StartDate"`
}

type CarrierParty struct {
	PartyIdentification ID    `xml:"cac:PartyIdentification"`
	PartyName           Name  `xml:"cac:PartyName"`
}

type TransportMeans struct {
	RoadTransportInstallation RoadInstallation `xml:"cac:RoadTransportInstallation"`
}

type RoadInstallation struct {
	LicensePlateID string `xml:"cbc:LicensePlateID"`
}

type Person struct {
	ID ID `xml:"cac:ID"`
}

type Delivery struct {
	Despatch      Despatch `xml:"cac:Despatch"`
	DeliveryAddress Address `xml:"cac:DeliveryAddress"`
}

type Despatch struct {
	DespatchAddress Address `xml:"cac:DespatchAddress"`
}

type Address struct {
	ID          string `xml:"cbc:ID"` // Ubigeo
	AddressLine Line   `xml:"cac:AddressLine"`
}

type Line struct {
	Line string `xml:"cbc:Line"`
}

type HandlingUnit struct {
	TransportEquipment []Equipment `xml:"cac:TransportEquipment"`
}

type Equipment struct {
	ID string `xml:"cbc:ID"`
}

type DespatchLine struct {
	ID               string   `xml:"cbc:ID"`
	DeliveredQuantity Measure  `xml:"cbc:DeliveredQuantity"`
	OrderLineReference Reference `xml:"cac:OrderLineReference,omitempty"`
	Item             Item     `xml:"cac:Item"`
}

type Reference struct {
	LineID string `xml:"cbc:LineID"`
}

type Item struct {
	Description string `xml:"cbc:Description"`
	ID          ID     `xml:"cac:SellersItemIdentification"`
}

// GreRequest represents the JSON payload for the new GRE REST API
type GreRequest struct {
	Archivo struct {
		NomArchivo string `json:"nomArchivo"`
		ArcGreZip  string `json:"arcGreZip"`
		HashZip    string `json:"hashZip"`
	} `json:"archivo"`
}

// GreResponse represents the response from the GRE REST API (POST)
type GreResponse struct {
	NumTicket string `json:"numTicket"`
	FecPedido string `json:"fecPedido"`
}

// GreStatusResponse represents the response from the GRE status query (GET)
type GreStatusResponse struct {
	CodRespuesta string `json:"codRespuesta"`
	ArcCdr       string `json:"arcCdr"`
	IndProg      string `json:"indProg"`
	Error        *struct {
		NumError string `json:"numError"`
		DesError string `json:"desError"`
	} `json:"error,omitempty"`
}

// OAuthToken represents the token response from SUNAT
type OAuthToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	IssuedAt    time.Time
}
