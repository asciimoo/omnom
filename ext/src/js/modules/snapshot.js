import {
    executeScriptToPromise,
    browser as br,
    walkDOM,
    fullURL,
    getSiteUrl,
    setSiteUrl
} from './utils';
import { downloadFile } from './file-download';
import { sanitizeCSS } from './sanitize';
import { saveBookmark } from "./main";

const nodeTransformFunctons = new Map([
    ['SCRIPT', (node) => node.remove()],
    ['LINK', transformLink],
    ['STYLE', transformStyle],
    ['IMG', transfromImg],
    ['BASE', setBaseUrl]
]);

const messageHandlers = new Map([
    ['pong', handlePongMessage],
    ['domData', handleDomDataMessage]
]);

const styleNodes = new Map();

let styleIndex = 0;
let commChan = null;
let numberOfIframes = 0;
let siteContent = {
    'main': null,
    'iframes': [],
};

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

function handlePongMessage(msg) {
    numberOfIframes += 1;
}

function handleDomDataMessage(msg) {
    if (msg.data.url == getSiteUrl()) {
        siteContent.main = msg.data;
    } else {
        siteContent.iframes.push(msg.data);
    }
    if (siteContent.iframes.length >= numberOfIframes) {
        parseDom();
    }
}

async function createSnapshot() {
    commChan.postMessage({type: "getDom"});
}

async function parseDom() {
    const doc = siteContent.main;
    const dom = transformDom(doc);
    saveBookmark({
        'dom': `${doc.doctype}${dom.outerHTML}`,
        'text': dom.getElementsByTagName("body")[0].innerText,
        'favicon': favicon
    });
}

async function transformDom(doc) {
    const dom = document.createElement('html');
    dom.innerHTML = doc.html;
    for (const k in doc.attributes) {
        dom.setAttribute(k, doc.attributes[k]);
    }
    await walkDOM(dom, transformNode);
    setStyleNodes(dom);
    let favicon = document.getElementById('favicon')?.value;
    if (!favicon) {
        favicon = await downloadFile(fullURL('/favicon.ico'));
        if (favicon) {
            const faviconElement = document.createElement('style');
            faviconElement.setAttribute('rel', 'icon');
            faviconElement.setAttribute('href', favicon);
            document.getElementsByTagName('head')[0].appendChild(faviconElement);
        }
    }
    return dom;
}

function setStyleNodes(dom) {
    const sortedStyles = new Map([...styleNodes.entries()].sort((e1, e2) => e1[0] - e2[0]));
    let parent;
    if (dom.getElementsByTagName("head")) {
        parent = dom.getElementsByTagName("head")[0];
    } else {
        parent = dom.documentElement;
    }
    sortedStyles.forEach(style => {
        parent.appendChild(style);
    });
}

async function transformNode(node) {
    if (node.nodeType !== Node.ELEMENT_NODE) {
        return;
    }
    const transformFunction = nodeTransformFunctons.get(node.nodeName);
    await rewriteAttributes(node);
    if (transformFunction) {
        await transformFunction(node);
        return;
    }
    return;
}

async function transformLink(node) {
    if (node.attributes.rel && node.attributes.rel.nodeValue.trim().toLowerCase() == 'stylesheet') {
        if (!node.attributes.href) {
            return;
        }
        const index = styleIndex++;
        const cssHref = fullURL(node.attributes.href.nodeValue);
        const style = document.createElement('style');
        const cssText = await downloadFile(cssHref);
        style.innerHTML = await sanitizeCSS(cssText, cssHref);
        styleNodes.set(index, style);
        node.remove();
    }
    if ((node.getAttribute('rel') || '').trim().toLowerCase() == 'icon' || (node.getAttribute('rel') || '').trim().toLowerCase() == 'shortcut icon') {
        const favicon = await downloadFile(node.href);
        document.getElementById('favicon').value = favicon;
        node.href = favicon;
    }
}

async function transformStyle(node) {
    const innerText = await sanitizeCSS(node.innerText, getSiteUrl());
    node.innerText = innerText;
}

async function transfromImg(node) {
    const src = await downloadFile(node.getAttribute('src'));
    node.src = src;
    if (node.getAttribute('srcset')) {
        let val = node.getAttribute('srcset');
        let newParts = [];
        for (let s of val.split(',')) {
            let srcParts = s.trim().split(' ');
            srcParts[0] = await downloadFile(srcParts[0]);
            newParts.push(srcParts.join(' '));
        }
        node.setAttribute('srcset', newParts.join(', '));
    }
}

async function setBaseUrl(node) {
    setSiteUrl(fullURL(node.getAttribute('href')));
}

async function rewriteAttributes(node) {
    const nodeAttributeArray = [...node.attributes];
    return Promise.allSettled(nodeAttributeArray.map(async (attr) => {
        if (attr.nodeName.startsWith('on') || attr.nodeValue.startsWith('javascript:')) {
            attr.nodeValue = '';
        }
        if (attr.nodeName == 'href') {
            attr.nodeValue = fullURL(attr.nodeValue);
        }
        if (attr.nodeName == 'style') {
            const sanitizedValue = await sanitizeCSS(`a{${attr.nodeValue}}`, getSiteUrl());
            attr.nodeValue = sanitizedValue.substr(4, sanitizedValue.length - 6);
        }
    }));
}

setupComms();

export { createSnapshot }
