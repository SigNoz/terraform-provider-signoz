package datasource

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// addErr adds an error to the diagnostics.
func addErr(diagnostics *diag.Diagnostics, err error, resource string) {
	if err == nil {
		return
	}

	diagnostics.AddError(
		fmt.Sprintf("failed to %s %s", operationRead, resource),
		err.Error(),
	)
}
