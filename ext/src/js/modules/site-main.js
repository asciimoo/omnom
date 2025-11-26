// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

/**
 * @fileoverview Content script for site integration and communication.
 * Handles message passing between the page and extension.
 */

import { getDomData } from "./get-dom-data";

/**
 * Map of message handlers for incoming messages
 * @type {Map<string, Function>}
 */
const messageHandlers = new Map([
    ['ping', handlePingMessage],
    ['getDom', handleGetDomMessage]
]);

/**
 * Initializes communication channels with the extension
 */
function initComms() {

    // messages from background.js
    chrome.runtime.onMessage.addListener(function(msg) {
        if(msg.action != "verify-settings-save") {
            console.log("Invalid message from background.js:", msg);
        }
        if(confirm("Do you want to use this account from your Omnom extension?")) {
            chrome.runtime.sendMessage({"action": "accept-settings"});
        } else {
            // TODO do not display this message again
            chrome.runtime.sendMessage({"action": "reject-settings"});
        }
    });

    // messages from popup.js
    chrome.runtime.onConnect.addListener(port => {
        port.onMessage.addListener((msg, commChan) => {
            // TODO generate static extension id and check the full ID, not just the schema
            if(!commChan.sender.origin.startsWith('chrome-extension://')) {
                console.log("invalid origin");
                return;
            }
            const msgHandler = messageHandlers.get(msg.type);
            if (msgHandler) {
                msgHandler(msg, commChan);
            } else {
                console.log("unknown message: ", msg);
            }
        });
    });
}

/**
 * Handles ping messages from the extension
 * @async
 * @param {Object} msg - The ping message
 * @param {Object} commChan - The communication channel
 */
async function handlePingMessage(msg, commChan) {
    commChan.postMessage({type: 'pong'});
    if(chrome.runtime.lastError) {
        console.log("Failed to deliver pong message", chrome.runtime.lastError);
    }
}

/**
 * Handles requests for DOM data from the extension
 * @async
 * @param {Object} msg - The DOM data request message
 * @param {Object} commChan - The communication channel
 */
async function handleGetDomMessage(msg, commChan) {
    commChan.postMessage({
        type: 'domData',
        data: getDomData(),
        isIframe: self != top
        //isIframe: self != top || document.location.ancestorOrigins.length
    });
    if(chrome.runtime.lastError) {
        console.log("Failed to deliver domData message", chrome.runtime.lastError);
    }
}

/**
 * Main entry point for site integration
 * Initializes the communication system
 */
export default function () {
    initComms();

}
