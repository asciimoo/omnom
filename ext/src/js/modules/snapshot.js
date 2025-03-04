// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

async function createSnapshot(doc) {
    await doc.transformDom();
    return {
        'dom': doc.getDomAsText(),
        // use ||| as a delimiter to separate semantically independent parts of the site's text
        'text': extractVisibleTextBlocks(doc.dom.getElementsByTagName("body")).join('|||'),
        'favicon': doc.favicon
    };
}

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

function extractVisibleTextBlocks(el) {

    const walker = document.createTreeWalker(el, NodeFilter.SHOW_TEXT | NodeFilter.SHOW_ELEMENT);
    let curTexts = [];
    let parentNode = el.tagName;
    let texts = [];
    let n = walker.nextNode();
    let visible = el.checkVisibility();
    while(n) {
        if(n.nodeType == Node.ELEMENT_NODE) {
            parentNode = n.tagName;
            visible = n.checkVisibility();
        } else if(n.nodeType == Node.TEXT_NODE) {
            if(!visible) {
                n = walker.nextNode();
                continue;
            }
            if(sectionElements.includes(parentNode)) {
                parentNode = undefined;
                if(curTexts.length > 0) {
                    let ct = curTexts.join('').replace(/\s+/g, " ").trim();
                    if(ct) {
                        texts.push(ct);
                    }
                    curTexts = [];
                }
            }
            const t = n.nodeValue;
            if(t) {
                curTexts.push(t);
            }
        }
        n = walker.nextNode();
    }

    if(curTexts.length > 0) {
        let ct = curTexts.join('').replace(/\s+/g, " ").trim();
        if(ct) {
            texts.push(ct);
        }
    }

    return texts;
}

export { createSnapshot }
