// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

function getDomData() {
    const sectionElements = [
        'ARTICLE',
        'ASIDE',
        'BLOCKQUOTE',
        'DIV',
        'DL',
        'DT',
        'FIGURE',
        'FOOTER',
        'H1',
        'H2',
        'H3',
        'H4',
        'H5',
        'H6',
        'LI',
        'MAIN',
        'NAV',
        'P',
        'SECTION',
        'TD',
        'TH',
    ];

    function skipInvisible(n) {
        if(n.nodeType != Node.ELEMENT_NODE) {
            return NodeFilter.FILTER_ACCEPT;
        }
        const style = window.getComputedStyle(n);
        const rect = n.getBoundingClientRect();
        // TODO calculate valid height by font size: rect.height < (style.fontSize.replace('px', '')/2))
        if(rect.width < 5 || rect.height < 5) {
            return NodeFilter.FILTER_REJECT;
        }
        if(style.display == 'none' || style.visibility == 'hidden' || style.opacity == '0') {
            return NodeFilter.FILTER_REJECT;
        }
        return NodeFilter.FILTER_ACCEPT;
    }

    function extractVisibleTextBlocks(el) {
        const walker = document.createTreeWalker(el, NodeFilter.SHOW_TEXT | NodeFilter.SHOW_ELEMENT, skipInvisible);
        let curTexts = [];
        let parentNode = el.tagName;
        let texts = [];
        while(walker.nextNode()) {
            let n = walker.currentNode;
            if(n.nodeType == Node.ELEMENT_NODE) {
                if(sectionElements.includes(n.tagName) && curTexts.length > 0) {
                    let ct = curTexts.join('').replace(/\s+/g, " ").trim();
                    if(ct) {
                        texts.push(ct);
                    }
                    curTexts = [];
                }
            } else if(n.nodeType == Node.TEXT_NODE) {
                curTexts.push(n.nodeValue);
            }
        }

        if(curTexts.length > 0) {
            let ct = curTexts.join('').replace(/\s+/g, " ").trim();
            if(ct) {
                texts.push(ct);
            }
        }

        return texts;
    }

    const html = document.documentElement;
    const styleElements = html.querySelectorAll('style');
    if (styleElements) {
        for (let style of styleElements) {
            let sheetRules;
            try {
                sheetRules = style.sheet?.cssRules;
            } catch (e) {
                console.log("failed to access to css, probably it comes from another extension: " + e);
                continue;
            }
            if (sheetRules) {
                const concatRules = [...sheetRules].reduce((rules, rule) => rules.concat(rule.cssText), '');
                style.innerText = concatRules;
            }
        }

    }
    const ret = {
        'html': html.cloneNode(true),
        'attributes': {},
        'title': '',
        'doctype': '',
        'iframeCount': html.querySelectorAll('iframe').length,
        'url': document.URL
    };
    // use ||| as a delimiter to separate semantically independent parts of the site's text
    ret.text = extractVisibleTextBlocks(document.body).join('|||');
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
    // check if iframe is populated by the page via js
    let iframes = html.querySelectorAll('iframe');
    if (iframes) {
        let iframeContents = [];
        for (let iframe of iframes) {
            const iframeDoc = iframe.contentDocument;
            if (iframeDoc) {
                // TODO handle <html> attributes & doctype
                iframeContents.push({
                    'html': btoa(unescape(encodeURIComponent(iframeDoc.documentElement.outerHTML))),
                    'url': iframeDoc.URL || document.URL
                });
            } else {
                iframeContents.push(0);
            }
        }
        let snapshotIframes = ret.html.querySelectorAll('iframe');
        for (let i in iframeContents) {
            if (!iframeContents[i]) {
                continue;
            }
            snapshotIframes[i].setAttribute('data-omnom-iframe-html', iframeContents[i]['html']);
            snapshotIframes[i].setAttribute('data-omnom-iframe-url', iframeContents[i]['url']);
        }
    }

    ret.html = ret.html.outerHTML;
    return ret;
}

export { getDomData }
