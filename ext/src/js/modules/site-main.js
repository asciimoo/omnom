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
        data: getDOMData()
    });
}

function getDOMData() {
    const html = document.getElementsByTagName('html')[0];
    const ret = {
        'html': html.cloneNode(true),
        'attributes': {},
        'title': '',
        'doctype': '',
        'iframeCount': document.getElementsByTagName('iframe').length,
        'url': document.URL
    };
    if (document.doctype) {
        ret.doctype = new XMLSerializer().serializeToString(document.doctype);
    }
    if (document.getElementsByTagName('title').length > 0) {
        ret.title = document.getElementsByTagName('title')[0].innerText;
    }
    [...html.attributes].forEach(attr => ret.attributes[attr.nodeName] = attr.nodeValue);
    let canvases = document.getElementsByTagName('canvas');
    if (canvases) {
        let canvasImages = [];
        for (let canvas of canvases) {
            let el = document.createElement("img");
            el.src = canvas.toDataURL();
            canvasImages.push(el);
        }
        let snapshotCanvases = ret.html.getElementsByTagName('canvas');
        for (let i in canvasImages) {
            let canvas = snapshotCanvases[i];
            canvas.parentNode.replaceChild(canvasImages[i], canvas);

        }
    }
    const styleElements = document.getElementsByTagName('style');
    if (styleElements) {
        for (let style of styleElements) {
            const sheetRules = style.sheet?.cssRules;
            if (sheetRules) {
                const concatRules = [...sheetRules].reduce((rules, rule) => rules.concat(rule.cssText), '');
                style.innerText = concatRules;
            }
        }

    }
    ret.html = ret.html.outerHTML;
    return ret;
}

export default function () {
    initComms();
}