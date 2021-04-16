// +build !js

package net

import (
	"net"
)

type (
	Dialer = net.Dialer
)

var Dial = net.Dial
