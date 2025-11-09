// pkg/contracts/version.go
package contracts

import "strings"

const (
	// SDKVersion define versão do SDK
	// BREAKING CHANGES incrementam major
	SDKVersion = "1.0.0"
)

// CompatibleWith verifica compatibilidade de plugin
// Compatível quando major version é igual
func CompatibleWith(pluginVersion string) bool {
	sdkParts := strings.Split(SDKVersion, ".")
	pluginParts := strings.Split(pluginVersion, ".")

	if len(sdkParts) == 0 || len(pluginParts) == 0 {
		return false
	}

	// Compatível quando major == major
	return sdkParts[0] == pluginParts[0]
}
