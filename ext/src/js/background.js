// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

"use strict";
chrome.runtime.onInstalled.addListener((details) => {
    console.log("Extension has been installed. Reason:", details.reason);
});
