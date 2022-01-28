import { downloadFile } from './file-download';
import { sanitizeCSS } from './sanitize';
import {
    UrlResolver,
    base64Decode,
    base64Encode
} from './utils';

class Document {
    constructor(html, url, doctype, htmlAttributes) {
        this.doctype = doctype;
        this.dom = document.createElement('html');
        this.iframes = [];
        this.favicon = null;
        this.dom.innerHTML = html;
        this.resolver = new UrlResolver(url);
        for (const k in htmlAttributes) {
            this.dom.setAttribute(k, htmlAttributes[k]);
        }
        this.nodeTransformFunctons = new Map([
            ['SCRIPT', (node) => node.remove()],
            ['LINK', this.transformLink],
            ['STYLE', this.transformStyle],
            ['IMG', this.transfromImg],
            ['IFRAME', this.transfromIframe],
            ['BASE', this.setUrl]
        ]);
    }

    absoluteUrl(url) {
        return this.resolver.resolve(url);
    }

    getDomAsText() {
        return `${this.doctype}${this.dom.outerHTML}`;
    }

    async transformDom() {
        await this.walkDOM(this.dom);
        if (!this.favicon) {
            this.favicon = await downloadFile(this.absoluteUrl('/favicon.ico'));
            if (this.favicon) {
                const faviconElement = document.createElement('link');
                faviconElement.setAttribute('rel', 'icon');
                faviconElement.setAttribute('href', this.favicon);
                this.dom.getElementsByTagName('head')[0].appendChild(faviconElement);
            }
        }
    }

    async walkDOM(node) {
        await this.transformNode(node);
        const children = [...node.childNodes];
        return Promise.allSettled(children.map(async (node) => {
            await this.walkDOM(node).catch(e => console.log("Error while transforming DOM: ", e));
        }));
    }

    async transformNode(node) {
        if (node.nodeType !== Node.ELEMENT_NODE) {
            return;
        }
        const transformFunction = this.nodeTransformFunctons.get(node.nodeName);
        if (transformFunction) {
            await transformFunction.call(this, node);
            return;
        }
        await this.rewriteAttributes(node);
        return;
    }

    async transformLink(node) {
        const rel = (node.getAttribute('rel') || '').trim().toLowerCase();
        switch (rel) {
            case 'stylesheet':
                if (!node.attributes.href) {
                    return;
                }
                const cssHref = this.absoluteUrl(node.attributes.href.nodeValue);
                const style = document.createElement('style');
                const cssText = await downloadFile(this.absoluteUrl(cssHref));
                style.innerHTML = await sanitizeCSS(cssText, cssHref);
                node.replaceWith(style);
                break;
            case 'icon':
            case 'shortcut icon':
            case 'apple-touch-icon':
            case 'apple-touch-icon-precomposed':
            case 'fluid-icon':
                const icon = await downloadFile(this.absoluteUrl(node.getAttribute('href')));
                node.setAttribute('href', icon);
                if (!this.favicon) {
                    this.favicon = icon;
                }
                break;
            case 'preconnect':
            case 'dns-prefetch':
                // TODO handle these elements more sophisticatedly
                node.setAttribute('href', '');
                break;
            case 'preload':
                const href = node.getAttribute('href');
                if (!href) {
                    break;
                }
                switch ((node.getAttribute('as') || '').toLowerCase()) {
                    case 'script':
                    case 'fetch':
                    case 'track':
                    case 'worker':
                        node.setAttribute('href', '');
                        break;
                    case 'font':
                        const inlineFont = await downloadFile(this.absoluteUrl(href));
                        if (inlineFont) {
                            node.setAttribute('href', inlineFont);
                        }
                        break;
                    case 'style':
                        const cssHref = this.absoluteUrl(href);
                        const style = document.createElement('style');
                        const cssText = await downloadFile(this.absoluteUrl(cssHref));
                        style.innerHTML = await sanitizeCSS(cssText, cssHref);
                        node.replaceWith(style);
                        break;
                    case 'document':
                    case 'embed':
                    case 'image':
                    case 'audio':
                    case 'object':
                        // TODO handle preloading of the above types
                        node.setAttribute('href', '');
                        break;

                }
                break;
        }
    }

    async transformStyle(node) {
        const innerText = await sanitizeCSS(node.innerText, this.absoluteUrl());
        node.innerText = innerText;
    }

    async transfromImg(node) {
        if (!node.getAttribute('src')) {
            return;
        }
        const src = await downloadFile(this.absoluteUrl(node.getAttribute('src')));
        node.src = src;
        if (node.getAttribute('srcset')) {
            let val = node.getAttribute('srcset');
            let newParts = [];
            for (let s of val.split(',')) {
                let srcParts = s.trim().split(' ');
                srcParts[0] = await downloadFile(this.absoluteUrl(srcParts[0]));
                newParts.push(srcParts.join(' '));
            }
            node.setAttribute('srcset', newParts.join(', '));
        }
    }

    async transfromIframe(node) {
        const dataHtmlAttr = 'data-omnom-iframe-html';
        const dataUrlAttr = 'data-omnom-iframe-url';
        if (node.hasAttribute(dataHtmlAttr)) {
            let iframeUrl = node.getAttribute(dataUrlAttr);
            if (iframeUrl == "about:blank") {
                iframeUrl = this.absoluteUrl();
            }
            let iframeHtml = base64Decode(node.getAttribute(dataHtmlAttr));
            let iframe = new Document(iframeHtml, iframeUrl, '<!DOCTYPE html>', {});
            await iframe.transformDom();
            console.log(iframe.getDomAsText());
            const inlineSrc = `data:text/html;base64,${base64Encode(iframe.getDomAsText())}`;
            node.setAttribute('src', inlineSrc);
            node.removeAttribute(dataHtmlAttr);
            node.removeAttribute(dataUrlAttr);
            return;
        }
        if (!node.getAttribute('src')) {
            return;
        }
        const src = this.absoluteUrl(node.getAttribute('src'));
        for (let iframe of this.iframes) {
            if (iframe.absoluteUrl() == src) {
                await iframe.transformDom();
                const inlineSrc = `data:text/html;base64,${base64Encode(iframe.getDomAsText())}`;
                node.setAttribute('src', inlineSrc);
                return;
            }
        }
        console.log("Meh, iframe not found: ", src);
        node.setAttribute('src', '');
    }

    async setUrl(node) {
        this.resolver.setBaseUrl(node.getAttribute('href'));
        node.setAttribute('href', '');
    }

    async rewriteAttributes(node) {
        const nodeAttributeArray = [...node.attributes];
        return Promise.allSettled(nodeAttributeArray.map(async (attr) => {
            if (attr.nodeName.startsWith('on') || attr.nodeValue.startsWith('javascript:')) {
                attr.nodeValue = '';
            }
            if (attr.nodeName == 'href') {
                attr.nodeValue = this.absoluteUrl(attr.nodeValue);
            }
            if (attr.nodeName == 'style') {
                const sanitizedValue = await sanitizeCSS(`a{${attr.nodeValue}}`, this.absoluteUrl());
                attr.nodeValue = sanitizedValue.substr(4, sanitizedValue.length - 6);
            }
        }));
    }
}

export { Document }
