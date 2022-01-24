function getDomData() {
    const html = document.getElementsByTagName('html')[0];
    const ret = {
        'html': html.cloneNode(true),
        'attributes': {},
        'title': '',
        'doctype': '',
        'iframeCount': html.querySelectorAll('iframe').length,
        'url': document.URL
    };
    if (document.doctype) {
        ret.doctype = new XMLSerializer().serializeToString(document.doctype);
    }
    if (document.getElementsByTagName('title').length > 0) {
        ret.title = document.getElementsByTagName('title')[0].innerText;
    }
    [...html.attributes].forEach(attr => ret.attributes[attr.nodeName] = attr.nodeValue);
    let canvases = html.querySelectorAll('canvas');
    if (canvases) {
        let canvasImages = [];
        for (let canvas of canvases) {
            let el = document.createElement("img");
            el.src = canvas.toDataURL();
            canvasImages.push(el);
        }
        let snapshotCanvases = ret.html.querySelectorAll('canvas');
        for (let i in canvasImages) {
            let canvas = snapshotCanvases[i];
            canvas.replaceWith(canvasImages[i]);

        }
    }
    const styleElements = html.querySelectorAll('style');
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

export { getDomData }
