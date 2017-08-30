// +build oracle

package config

import (
	"github.com/ory/hydra-oracle-plugin/plugin"
)

/**
 * This exists to force the oracle plugin to load from vendor directory
 * which will cause it to register itself as a vendored plugin (see backed_vendor & plugin/register.go)
 * this file will be excluded based on the build tag
 */

func init() {
	//register the plugin with hydra
	RegisterVendoredPlugin("oracle", plugin.OracleVendoredPlugin{}, nil)
}
