package resource

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// addErr adds an error to the diagnostics.
func addErr(diagnostics *diag.Diagnostics, err error, operation string, resource string) {
	if err == nil {
		return
	}

	diagnostics.AddError(
		fmt.Sprintf("failed to %s %s", operation, resource),
		err.Error(),
	)
}
