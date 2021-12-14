import { executeScriptToPromise, browser as br, walkDOM, downloadStatus, fullURL, inlineFile, getSiteUrl } from './utils';
import { sanitizeCSS } from './sanitize';

const nodeTransformFunctons = new Map([
    ['SCRIPT', (node) => node.remove()],
    ['LINK', transformLink],
    ['STYLE', transformStyle],
    ['IMG', transfromImg]
]);

let styleIndex = 0;

const styleNodes = new Map();

async function createSnapshot() {
    const doc = await getDOM();
    const dom = document.createElement('html');
    dom.innerHTML = doc.html;
    for (const k in doc.attributes) {
        dom.setAttribute(k, doc.attributes[k]);
    }
    await walkDOM(dom, transformNode);
    setStyleNodes(dom);
    if (!document.getElementById('favicon').value) {
        const favicon = await inlineFile(fullURL('/favicon.ico'));
        if (favicon) {
            document.getElementById('favicon').value = favicon;
            const faviconElement = document.createElement('style');
            faviconElement.setAttribute('rel', 'icon');
            faviconElement.setAttribute('href', favicon);
            document.getElementsByTagName('head')[0].appendChild(faviconElement);
        }
    }
    return {
        'dom': `${doc.doctype}${dom.outerHTML}`,
        'text': dom.getElementsByTagName("body")[0].innerText,
    };
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

async function getDOM() {
    const data = await executeScriptToPromise(getDOMData, br);
    if (data && data[0]) {
        return Promise.resolve(data[0]);
    } else {
        return Promise.reject('meh')
    }
}

function getDOMData() {
    const html = document.getElementsByTagName('html')[0];
    const ret = {
        'html': html.cloneNode(true),
        'attributes': {},
        'title': '',
        'doctype': '',
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
    ret.html = ret.html.outerHTML;
    return ret;
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
        const cssText = await inlineFile(cssHref);
        style.innerHTML = await sanitizeCSS(cssText, cssHref);
        styleNodes.set(index, style);
        node.remove();
    }
    if ((node.getAttribute('rel') || '').trim().toLowerCase() == 'icon' || (node.getAttribute('rel') || '').trim().toLowerCase() == 'shortcut icon') {
        const favicon = await inlineFile(node.href);
        document.getElementById('favicon').value = favicon;
        node.href = favicon;
    }
}

async function transformStyle(node) {
    const innerText = await sanitizeCSS(node.innerText, getSiteUrl());
    node.innerText = innerText;
}

async function transfromImg(node) {
    const src = await inlineFile(node.getAttribute('src'));
    node.src = src;
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
        if (attr.nodeName == 'srcset') {
            let newParts = [];
            for (let s of attr.nodeValue.split(',')) {
                let srcParts = s.trim().split(' ');
                srcParts[0] = await inlineFile(srcParts[0]);
                newParts.push(srcParts.join(' '));
            }
            attr.nodeValue = newParts.join(', ');
        }
    }));
}

export { createSnapshot, transformLink, transformStyle, transfromImg }