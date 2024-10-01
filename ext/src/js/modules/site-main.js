// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

import { getDomData } from "./get-dom-data";

const messageHandlers = new Map([
    ['ping', handlePingMessage],
    ['getDom', handleGetDomMessage]
]);

function initComms() {
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

async function handlePingMessage(msg, commChan) {
    commChan.postMessage({type: 'pong'});
    if(chrome.runtime.lastError) {
        console.log("Failed to deliver pong message", chrome.runtime.lastError);
    }
}

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

export default function () {
    initComms();
}
