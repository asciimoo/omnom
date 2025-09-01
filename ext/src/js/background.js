// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

"use strict";

import { getOmnomSettings } from './modules/utils';

let pending_settings = null;

const siteMessageHandlers = new Map([
    ['ping', handlePing],
    ['set-settings', handleSetSettings],
]);

const cjsMessageHandlers = new Map([
    ['accept-settings', handleAcceptSettings],
    ['reject-settings', handleRejectSettings],
]);

function siteMsgHandler(msg, sender, send) {
    const msgHandler = siteMessageHandlers.get(msg.action);
    if(msgHandler) {
        msgHandler(msg, sender, send);
    } else {
        console.log("unknown message: ", msg);
    }
}

function cjsMsgHandler(msg, sender) {
    const msgHandler = cjsMessageHandlers.get(msg.action);
    if(msgHandler) {
        msgHandler(msg, sender);
    } else {
        console.log("unknown message: ", msg);
    }
}

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

function handleSetSettings(msg, sender, send) {
    if(pending_settings) {
        send("settings pending");
        return;
    }
    msg['send'] = send;
    pending_settings = msg;
    chrome.tabs.sendMessage(sender.tab.id, {"action": "verify-settings-save"});
}

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
