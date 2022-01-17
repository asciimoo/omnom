import {
    executeScriptToPromise,
    browser as br,
    getSiteUrl
} from './utils';
import { saveBookmark } from "./main";
import { Document } from "./document";

const messageHandlers = new Map([
    ['pong', handlePongMessage],
    ['domData', handleDomDataMessage]
]);

let commChan = null;
let numberOfPages = 0;
let doc = null;
let iframes = [];

function setupComms() {
    br.tabs.query({
        active: true,
        currentWindow: true
    }, tabs => {
        let tab = tabs[0];
        commChan = br.tabs.connect(
            tab.id,
            {name: "omnom"}
        );
        commChan.onMessage.addListener((msg) => {
            const msgHandler = messageHandlers.get(msg.type);
            if(msgHandler) {
                msgHandler(msg);
            } else {
                console.log("unknown message: ", msg);
            }
            return true;
        });
        commChan.postMessage({type: "ping"});
    });
}

async function handlePongMessage(msg) {
    numberOfPages += 1;
}

async function handleDomDataMessage(msg) {
    let d = new Document(msg.data.html, msg.data.url, msg.data.doctype, msg.data.attributes);
    if (msg.data.url == getSiteUrl()) {
        doc = d;
    } else {
        iframes.push(d);
    }
    if (doc && iframes.length == numberOfPages -1) {
        doc.iframes = iframes;
        parseDom();
    }
}

async function createSnapshot() {
    commChan.postMessage({type: "getDom"});
}

async function parseDom() {
    await doc.transformDom();
    saveBookmark({
        'dom': doc.getDomAsText(),
        'text': doc.dom.getElementsByTagName("body")[0].innerText,
        'favicon': doc.favicon
    });
}

setupComms();

export { createSnapshot }
