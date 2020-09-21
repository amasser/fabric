/*
	Copyright NetFoundry, Inc.

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

/*
Package xweb provides facilities to creating composable xweb.WebHandlers and http.Server's from configuration files.

Basics

xweb provides customizable and extendable components to stand up multiple http.Server's listening on one or more
network interfaces.

Each xweb.Xweb is responsible for defining configuration sections to be parsed, parsing the configuration, starting
servers, and shutting down relevant server. An example implementation is included in the package: xweb.XwebImpl. This
implementation should cover most use cases. In addition xweb.XwebImpl makes use of xweb.Config which is reusable
component for parsing xweb.XwebImpl configuration sections. Both xweb.Xweb xweb.Config assume that configuration will
be acquired from some source and be presented as a map of interface{}-to-interface{} values
(map[interface{}]interface{}) per the openziti/foundation config.Subconfig pattern.

xweb.Config configuration sections allow the definition of an array of xweb.WebListener. In turn each xweb.WebListener
can listen on many interface/port combinations specified by an array of xweb.BindPoint and host many http.Handler's
by defining an array of xweb.API's that are converted into xweb.WebHandler's. xweb.WebHandler's are http.Handler's with
meta data and can be as complex or as simple as necessary - using other libraries or only the standard http Go
capabilities.

To deal with a single xweb.WebListener hosting multiple APIs as web.WebListener's, incoming requests must be forwarded
to the correct xweb.WebHandler. The responsibility is handled by another configurable http.Handler called an
"xweb demux handler". This handler's responsibility is to inspect incoming requests and forward them to the correct
xweb.WebHandler. It is specified by an xweb.DemuxFactory and a reference implementation, xweb.PathPrefixDemuxFactory
has been provided.
*/
package xweb

import (
	"context"
	"github.com/michaelquigley/pfxlog"
	"time"
)

// Implements config.Subconfig to allow Xweb implementations to be used during the normal Ziti component  startup
// and configuration phase.
type Xweb interface {
	Enabled() bool
	LoadConfig(cfgmap map[interface{}]interface{}) error
	Run()
	Shutdown()
}

const (
	DefaultIdentitySection = "identity"
	WebSection             = "web"
)

// A simple implementation of xweb.XWeb used for registration and configuration from controller.Controller.
// Implements necessary interfaces to be a config.Subconfig.
type XwebImpl struct {
	Config       *Config
	servers      []*Server
	Registry     WebHandlerFactoryRegistry
	DemuxFactory DemuxFactory
}

func NewXwebImpl(registry WebHandlerFactoryRegistry) *XwebImpl {
	return &XwebImpl{
		Registry:     registry,
		DemuxFactory: &PathPrefixDemuxFactory{},
		Config: &Config{
			DefaultIdentitySection: DefaultIdentitySection,
			WebSection:             WebSection,
		},
	}
}

// Whether this subconfig should be considered enabled
func (xwebimpl *XwebImpl) Enabled() bool {
	return xwebimpl.Config.Enabled()
}

// Handle subconfig operations for xweb.Xweb components
func (xwebimpl *XwebImpl) LoadConfig(cfgmap map[interface{}]interface{}) error {
	if err := xwebimpl.Config.Parse(cfgmap); err != nil {
		return err
	}

	//validate sets enabled flag to true on success
	if err := xwebimpl.Config.Validate(xwebimpl.Registry); err != nil {
		return err
	}

	return nil
}

// Starts the necessary xweb.Server's
func (xwebimpl *XwebImpl) Run() {
	for _, webListener := range xwebimpl.Config.WebListeners {
		server, err := NewServer(webListener, xwebimpl.DemuxFactory, xwebimpl.Registry)

		if err != nil {
			pfxlog.Logger().Fatalf("error starting xweb server for %s: %v", webListener.Name, err)
		}

		xwebimpl.servers = append(xwebimpl.servers, server)

		if err := server.Start(); err != nil {
			pfxlog.Logger().Errorf("error starting xgweb_rest server %s: %v", webListener.Name, err)
		}
	}
}

// Stop all running xweb.Server's
func (xwebimpl *XwebImpl) Shutdown() {
	for _, server := range xwebimpl.servers {
		localServer := server
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
			defer cancel()
			localServer.Shutdown(ctx)
		}()

	}
}
