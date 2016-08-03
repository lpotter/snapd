// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package builtin

import (
	"github.com/snapcore/snapd/interfaces"
)

const sensorsConnectedPlugAppArmor = `

/dev/spidev* rw,
/dev/i2c* rw,
/dev/input/event* rw,
/sys/bus/i2c/devices/** rw,
/sys/devices/** rw,
/var/run/* rw,

`

const sensorsConnectedPlugSecComp = `
# Description: Can access sockets.
# Usage: common
accept
accept4
bind
connect
getpeername
getsockname
getsockopt
listen
recv
recvfrom
recvmmsg
recvmsg
send
sendmmsg
sendmsg
sendto
setsockopt
shutdown

# LP: #1446748 - limit this to AF_INET/AF_INET6
socket

# This is an older interface and single entry point that can be used instead
# of socket(), bind(), connect(), etc individually. While we could allow it,
# we wouldn't be able to properly arg filter socketcall for AF_INET/AF_INET6
# when LP: #1446748 is implemented.
socketcall
`

// NewSensorsInterface returns a new "sensors" interface.
func NewSensorsInterface() interfaces.Interface {
	return &commonInterface{
		name: "sensors",
		connectedPlugAppArmor: sensorsConnectedPlugAppArmor,
		connectedPlugSecComp:  sensorsConnectedPlugSecComp,
		reservedForOS:         true,
	}
}
