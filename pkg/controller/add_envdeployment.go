package controller

import (
	"github.com/alexeykazakov/devconsole/pkg/controller/envdeployment"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, envdeployment.Add)
}
