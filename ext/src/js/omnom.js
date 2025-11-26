// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

/**
 * @fileoverview Main entry point for the Omnom extension popup.
 * Initializes the popup interface when the DOM is ready.
 */

import { displayPopup } from "./modules/main";

document.addEventListener('DOMContentLoaded', displayPopup);
