/*
	(c) Copyright NetFoundry, Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package transwarp

import (
	"fmt"
	"github.com/netfoundry/ziti-foundation/transport"
	"io"
	"net"
)

func Listen(bindAddress, name string, incoming chan transport.Connection) (io.Closer, error) {
	listenAddress, err := net.ResolveUDPAddr("udp", bindAddress)
	if err != nil {
		return nil, fmt.Errorf("error resolving bind address (%w)", err)
	}

	socket, err := net.ListenUDP("udp", listenAddress)
	if err != nil {
		return nil, fmt.Errorf("error listening (%w)", err)
	}

	incoming <- newConnection(&transport.ConnectionDetail{
		Address: "transwarp" + bindAddress,
		InBound: true,
		Name:    name,
	}, socket)

	return socket, nil
}