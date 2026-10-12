package statebinding

import (
	"talenro.local/platform/internal/nodecontrol/statecomposition"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"talenro.local/platform/internal/nodecontrol/statematerials"
)

type Input struct {
	Main     statecontract.ModelInput
	Previous statematerials.Input
	Proposed statecomposition.Input
}
