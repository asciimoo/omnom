// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+
import {
   extractVisibleTextBlocks,
} from './utils';

function getDomData() {
    function createShadowDomTemplates(node) {
		let ni = document.createNodeIterator(node, NodeFilter.SHOW_ELEMENT)
		let n;
		let elCount = 0;
		let tpls = [];
		while((n = ni.nextNode())) {
			if(!n.shadowRoot) {
				continue;
			}
			let cHTML = [];
			for(let c of n.shadowRoot.children) {
				cHTML.push(c.outerHTML);
			}
			n.setAttribute("omnomshadowroot", elCount++);
			let tpl = document.createElement('template');
			tpl.innerHTML = cHTML.join('');
			tpls.push(tpl);
		}
		return tpls;
	}

	function applyShadowDomTemplates(node, tpls) {
		node.querySelectorAll("[omnomshadowroot]").forEach(el => {
			let idx = Number(el.getAttribute("omnomshadowroot"));
			el.prepend(tpls[idx]);
		});
	}

    const shadowTpls = createShadowDomTemplates(document.getRootNode());
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
                style.textContent = concatRules;
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
    applyShadowDomTemplates(ret.html, shadowTpls);
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
