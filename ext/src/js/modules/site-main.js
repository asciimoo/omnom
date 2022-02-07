import { getDomData } from "./get-dom-data";

const messageHandlers = new Map([
    ['ping', handlePingMessage],
    ['getDom', handleGetDomMessage]
]);

function initComms() {
    chrome.runtime.onConnect.addListener(port => {
        port.onMessage.addListener((msg, commChan) => {
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
}

async function handleGetDomMessage(msg, commChan) {
    commChan.postMessage({
        type: 'domData',
        data: getDomData(),
        isIframe: self != top
        //isIframe: self != top || document.location.ancestorOrigins.length
    });
}

export default function () {
    initComms();
}
