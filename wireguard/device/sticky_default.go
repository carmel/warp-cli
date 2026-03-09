//go:build !linux

package device

import (
	"github.com/carmel/warp-cli/wireguard/conn"
	"github.com/carmel/warp-cli/wireguard/rwcancel"
)

func (device *Device) startRouteListener(_ conn.Bind) (*rwcancel.RWCancel, error) {
	return nil, nil
}
