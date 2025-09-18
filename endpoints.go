// Package sunatlib defines SUNAT service endpoints
package sunatlib

// SUNAT Production Endpoints
const (
	// Production endpoints for electronic invoicing
	SUNATProductionBillService     = "https://e-factura.sunat.gob.pe/ol-ti-itcpfegem/billService"
	SUNATProductionRetentionService = "https://e-factura.sunat.gob.pe/ol-ti-itemision-otroscpe-gem/billService"
	SUNATProductionGuideService    = "https://e-factura.sunat.gob.pe/ol-ti-itemision-guia-gem/billService"

	// Production endpoint for document validation
	SUNATProductionValidationService = "https://e-factura.sunat.gob.pe/ol-it-wsconsvalidcpe/billValidService"
)

// SUNAT Beta/Testing Endpoints
const (
	// Beta endpoints for electronic invoicing (testing)
	SUNATBetaBillService     = "https://e-beta.sunat.gob.pe/ol-ti-itcpfegem-beta/billService"
	SUNATBetaRetentionService = "https://e-beta.sunat.gob.pe/ol-ti-itemision-otroscpe-gem-beta/billService"
	SUNATBetaGuideService    = "https://e-beta.sunat.gob.pe/ol-ti-itemision-guia-gem-beta/billService"

	// Beta endpoint for document validation (testing)
	SUNATBetaValidationService = "https://e-beta.sunat.gob.pe/ol-it-wsconsvalidcpe/billValidService"
)

// Environment types
type Environment int

const (
	Production Environment = iota
	Beta
)

// GetBillServiceEndpoint returns the appropriate billService endpoint based on environment
func GetBillServiceEndpoint(env Environment) string {
	switch env {
	case Beta:
		return SUNATBetaBillService
	default:
		return SUNATProductionBillService
	}
}

// GetValidationServiceEndpoint returns the appropriate validation service endpoint based on environment
func GetValidationServiceEndpoint(env Environment) string {
	switch env {
	case Beta:
		return SUNATBetaValidationService
	default:
		return SUNATProductionValidationService
	}
}

// GetRetentionServiceEndpoint returns the appropriate retention/perception service endpoint based on environment
func GetRetentionServiceEndpoint(env Environment) string {
	switch env {
	case Beta:
		return SUNATBetaRetentionService
	default:
		return SUNATProductionRetentionService
	}
}

// GetGuideServiceEndpoint returns the appropriate guide service endpoint based on environment
func GetGuideServiceEndpoint(env Environment) string {
	switch env {
	case Beta:
		return SUNATBetaGuideService
	default:
		return SUNATProductionGuideService
	}
}