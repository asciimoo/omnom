// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

/**
 * @fileoverview Background script for the Omnom browser extension.
 * Handles message passing between the extension, content scripts, and external sites.
 */

"use strict";

import { getOmnomSettings } from './modules/utils';

/**
 * Pending settings object awaiting user confirmation
 * @type {Object|null}
 */
let pending_settings = null;

/**
 * Map of message handlers for external site messages
 * @type {Map<string, Function>}
 */
const siteMessageHandlers = new Map([
    ['ping', handlePing],
    ['set-settings', handleSetSettings],
]);

/**
 * Map of message handlers for content script messages
 * @type {Map<string, Function>}
 */
const cjsMessageHandlers = new Map([
    ['accept-settings', handleAcceptSettings],
    ['reject-settings', handleRejectSettings],
]);

/**
 * Handles messages from external sites
 * @param {Object} msg - The message object
 * @param {Object} sender - The message sender
 * @param {Function} send - Response callback function
 */
function siteMsgHandler(msg, sender, send) {
    const msgHandler = siteMessageHandlers.get(msg.action);
    if(msgHandler) {
        msgHandler(msg, sender, send);
    } else {
        console.log("unknown message: ", msg);
    }
}

/**
 * Handles messages from content scripts
 * @param {Object} msg - The message object
 * @param {Object} sender - The message sender
 */
function cjsMsgHandler(msg, sender) {
    const msgHandler = cjsMessageHandlers.get(msg.action);
    if(msgHandler) {
        msgHandler(msg, sender);
    } else {
        console.log("unknown message: ", msg);
    }
}

/**
 * Handles ping messages from external sites to verify extension settings
 * @param {Object} msg - The ping message
 * @param {Object} sender - The message sender
 * @param {Function} send - Response callback function
 */
function handlePing(msg, sender, send) {
    if(!sender.url.startsWith(msg.url)) {
        return;
    }
    getOmnomSettings((data) => {
        let resp = {"action": "pong"};
        if(!data.omnom_url) {
            resp.url = "empty";
        } else if(data.omnom_url == msg.url) {
            resp.url = "same";
        } else {
            resp.url = "different";
        }
        send(resp);
    });
}

/**
 * Handles requests to update extension settings
 * @param {Object} msg - The settings message containing url and token
 * @param {Object} sender - The message sender
 * @param {Function} send - Response callback function
 */
function handleSetSettings(msg, sender, send) {
    msg['send'] = send;
    pending_settings = msg;
    chrome.tabs.query({ active: true, currentWindow: true }, function (tabs) {
        console.log(tabs[0].id);
    });
    chrome.tabs.sendMessage(sender.tab.id, {"action": "verify-settings-save"})
        .then(resp => {
            console.log("settings response received from page context", msg, resp);
        })
        .catch(err => console.log("failed to receive msg from page:", err, msg, sender.tab));
}

/**
 * Handles acceptance of pending settings changes
 * @param {Object} msg - The acceptance message
 * @param {Object} sender - The message sender
 */
function handleAcceptSettings(msg, sender) {
    if(!pending_settings) {
        return;
    }
    chrome.storage.local.set({
        omnom_url: pending_settings.url,
        omnom_token: pending_settings.token,
    }).then(() => {
        pending_settings.send("success");
        pending_settings = null;
    });
}

/**
 * Handles rejection of pending settings changes
 * @param {Object} msg - The rejection message
 * @param {Object} sender - The message sender
 */
function handleRejectSettings(msg, sender) {
    if(!pending_settings) {
        return;
    }
    pending_settings.send("rejected");
    pending_settings = null;
}

chrome.runtime.onInstalled.addListener((details) => {
    console.log("Extension has been installed. Reason:", details.reason);
});

// site
chrome.runtime.onMessageExternal.addListener(siteMsgHandler);
// content.js (site-main.js)
chrome.runtime.onMessage.addListener(cjsMsgHandler);
