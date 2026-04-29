// Package sunatlib defines SUNAT service endpoints
package sunatlib

import "fmt"

// SUNAT Production Endpoints
const (
	// Production endpoints for electronic invoicing
	SUNATProductionBillService     = "https://e-factura.sunat.gob.pe/ol-ti-itcpfegem/billService"
	SUNATProductionRetentionService = "https://e-factura.sunat.gob.pe/ol-ti-itemision-otroscpe-gem/billService"
	SUNATProductionGuideService    = "https://e-factura.sunat.gob.pe/ol-ti-itemision-guia-gem/billService"

	// Production endpoint for document validation
	SUNATProductionValidationService = "https://e-factura.sunat.gob.pe/ol-it-wsconsvalidcpe/billValidService"

	// New GRE REST and OAuth Endpoints
	SUNATProductionGREToken = "https://api-seguridad.sunat.gob.pe/v1/clientessol/%s/oauth2/token"
	SUNATProductionGREApi   = "https://api-cpe.sunat.gob.pe/v1/contribuyente/gem/comprobantes"
)

// SUNAT Beta/Testing Endpoints
const (
	// Beta endpoints for electronic invoicing (testing)
	SUNATBetaBillService     = "https://e-beta.sunat.gob.pe/ol-ti-itcpfegem-beta/billService"
	SUNATBetaRetentionService = "https://e-beta.sunat.gob.pe/ol-ti-itemision-otroscpe-gem-beta/billService"
	SUNATBetaGuideService    = "https://e-beta.sunat.gob.pe/ol-ti-itemision-guia-gem-beta/billService"

	// Beta endpoint for document validation (testing)
	SUNATBetaValidationService = "https://e-beta.sunat.gob.pe/ol-it-wsconsvalidcpe/billValidService"

	// Beta GRE REST and OAuth Endpoints
	// NOTE: api-cpe-beta.sunat.gob.pe often has DNS resolution issues.
	// Workaround: Add "161.132.21.169 api-cpe-beta.sunat.gob.pe" to your /etc/hosts file.
	SUNATBetaGREToken = "https://api-seguridad.sunat.gob.pe/v1/clientessol/%s/oauth2/token"
	SUNATBetaGREApi   = "https://api-cpe-beta.sunat.gob.pe/v1/contribuyente/gem/comprobantes"

	// NubeFact GRE Sandbox Endpoints (Alternative for testing)
	NubeFactGREToken = "https://gre-test.nubefact.com/v1/clientessol/%s/oauth2/token"
	NubeFactGREApi   = "https://gre-test.nubefact.com/v1/contribuyente/gem/comprobantes"
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

// GetGRETokenEndpoint returns the appropriate GRE OAuth token endpoint based on environment
func GetGRETokenEndpoint(env Environment, clientID string) string {
	endpoint := SUNATProductionGREToken
	if env == Beta {
		endpoint = SUNATBetaGREToken
	}
	return fmt.Sprintf(endpoint, clientID)
}

// GetGREApiEndpoint returns the appropriate GRE REST API endpoint based on environment
func GetGREApiEndpoint(env Environment) string {
	if env == Beta {
		return SUNATBetaGREApi
	}
	return SUNATProductionGREApi
}