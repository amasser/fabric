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

package xt_random

import (
	"github.com/netfoundry/ziti-fabric/controller/xt"
	"math/rand"
)

/**
The random strategy uses a pure random selection from available terminators. It does not take an costs/weights
or precedences into account. A slightly smarter implmementation would take precendence into account. However,
for now this implementation exists a reference point that we can compare other strategies to.
*/

func NewFactory() xt.Factory {
	return factory{}
}

type factory struct{}

func (f factory) GetStrategyName() string {
	return "random"
}

func (f factory) NewStrategy() xt.Strategy {
	return strategy{}
}

type strategy struct{}

func (s strategy) Select(terminators []xt.WeightedTerminator, totalWeight uint32) (xt.Terminator, error) {
	count := len(terminators)
	if count == 1 {
		return terminators[0], nil
	}
	selected := rand.Intn(count)
	return terminators[selected], nil
}

func (s strategy) NotifyEvent(xt.TerminatorEvent) {}

func (s strategy) HandleTerminatorChange(xt.StrategyChangeEvent) error {
	return nil
}