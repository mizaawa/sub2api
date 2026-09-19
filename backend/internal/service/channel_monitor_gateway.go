package service

import (
	"net"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

// channelMonitorInternalGatewayEndpoint returns the backend listener address
// used by managed monitors. The credential is valid only for this gateway, so
// it must never be sent to an administrator-supplied external endpoint.
func channelMonitorInternalGatewayEndpoint(cfg *config.Config) string {
	host := "127.0.0.1"
	port := 8080
	if cfg != nil {
		configuredHost := strings.TrimSpace(cfg.Server.Host)
		switch configuredHost {
		case "", "0.0.0.0", "::", "[::]":
		default:
			host = strings.Trim(configuredHost, "[]")
		}
		if cfg.Server.Port > 0 {
			port = cfg.Server.Port
		}
	}
	return "http://" + net.JoinHostPort(host, strconv.Itoa(port))
}
