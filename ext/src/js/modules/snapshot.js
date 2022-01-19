import {
    executeScriptToPromise,
    browser as br,
    getSiteUrl
} from './utils';
import { saveBookmark } from "./main";
import { Document } from "./document";
import { getDomData } from "./get-dom-data";

const messageHandlers = new Map([
    ['pong', handlePongMessage],
    ['domData', handleDomDataMessage]
]);

let commChan = null;
let numberOfPages = 0;
let doc = null;
let iframes = [];

function setupComms(msg) {
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
        if (msg) {
            commChan.postMessage(msg);
        }
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
        processDom(doc);
    }
}

async function createSnapshot() {
    if (!numberOfPages) {
        console.log("content js isn't running, falling back to the naive snapshotting, which does not include iframes");
        let data = await executeScriptToPromise(getDomData, br);
        if (!data) {
            // TODO display error to user
            console.log("failed to get dom information");
            return;
        }
        data = data[0];
        doc = new Document(data.html, data.url, data.doctype, data.attributes);
        processDom(doc);
        return
    }
    commChan.postMessage({type: "getDom"});
}

async function processDom(doc) {
    await doc.transformDom();
    saveBookmark({
        'dom': doc.getDomAsText(),
        'text': doc.dom.getElementsByTagName("body")[0].innerText,
        'favicon': doc.favicon
    });
}

setupComms({type: "ping"});

export { createSnapshot }
