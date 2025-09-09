package sunatlib

import "fmt"

// ConsultationClient handles RUC and DNI consultation services independently
type ConsultationClient struct {
	rucService *RUCService
	dniService *DNIService
}

// NewConsultationClient creates a new consultation client with both services
func NewConsultationClient(decolectaAPIKey string) *ConsultationClient {
	return &ConsultationClient{
		rucService: NewRUCService(decolectaAPIKey),
		dniService: NewDNIService(),
	}
}

// NewRUCConsultationClient creates a client only for RUC consultation
func NewRUCConsultationClient(decolectaAPIKey string) *ConsultationClient {
	return &ConsultationClient{
		rucService: NewRUCService(decolectaAPIKey),
	}
}

// NewDNIConsultationClient creates a client only for DNI/CE consultation (free)
func NewDNIConsultationClient() *ConsultationClient {
	return &ConsultationClient{
		dniService: NewDNIService(),
	}
}

// ConsultRUC performs a basic RUC consultation
func (c *ConsultationClient) ConsultRUC(ruc string) (*RUCBasicResponse, error) {
	if c.rucService == nil {
		return nil, fmt.Errorf("RUC service not available - use NewConsultationClient() or NewRUCConsultationClient()")
	}
	return c.rucService.ConsultBasic(ruc)
}

// ConsultRUCFull performs a complete RUC consultation
func (c *ConsultationClient) ConsultRUCFull(ruc string) (*RUCFullResponse, error) {
	if c.rucService == nil {
		return nil, fmt.Errorf("RUC service not available - use NewConsultationClient() or NewRUCConsultationClient()")
	}
	return c.rucService.ConsultFull(ruc)
}

// ConsultDNI performs a DNI consultation
func (c *ConsultationClient) ConsultDNI(dni string) (*DNIResponse, error) {
	if c.dniService == nil {
		return nil, fmt.Errorf("DNI service not available - use NewConsultationClient() or NewDNIConsultationClient()")
	}
	return c.dniService.ConsultDNI(dni)
}

// ConsultCE performs a Carnet de Extranjería consultation
func (c *ConsultationClient) ConsultCE(ce string) (*DNIResponse, error) {
	if c.dniService == nil {
		return nil, fmt.Errorf("DNI service not available - use NewConsultationClient() or NewDNIConsultationClient()")
	}
	return c.dniService.ConsultCE(ce)
}