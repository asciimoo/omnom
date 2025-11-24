// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package model

import (
	"fmt"
	"strings"
)

// CreateGlob converts a search string into a SQL LIKE pattern.
func CreateGlob(s string) string {
	if strings.Contains(s, "*") {
		return strings.ReplaceAll(s, "*", "%")
	}
	return fmt.Sprintf("%%%s%%", s)
}
