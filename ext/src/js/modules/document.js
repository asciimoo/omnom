// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

/**
 * @fileoverview Document class for processing and transforming HTML documents.
 * Handles resource extraction, sanitization, and DOM transformations for snapshotting.
 */

import { downloadFile } from './file-download';
import { ResourceStorage } from './resources';
import { Sanitizer } from './sanitize';
import {
    UrlResolver,
    base64Decode,
    base64Encode
} from './utils';

/**
 * Represents an HTML document with transformation capabilities
 * @class
 */
class Document {
    /**
     * Creates a Document instance
     * @param {string} html - The HTML content
     * @param {string} text - The extracted text content
     * @param {string} url - The document URL
     * @param {string} doctype - The document type declaration
     * @param {string} title - The document title
     * @param {Object} htmlAttributes - Attributes from the HTML element
     */
    constructor(html, text, url, doctype, title, htmlAttributes) {
        this.doctype = doctype;
        this.dom = document.createElement('html');
        this.iframes = [];
        this.favicon = null;
        this.dom.innerHTML = html;
        this.originalLength = html.length;
        this.resolver = new UrlResolver(url);
        this.resources = new ResourceStorage();
        this.sanitizer = new Sanitizer(this.resources);
        this.multimediaCount = 0;
        this.text = text;
        for (const k in htmlAttributes) {
            this.dom.setAttribute(k, htmlAttributes[k]);
        }
        this.nodeTransformFunctions = new Map([
            ['SCRIPT', (node) => node.remove()],
            ['TEMPLATE', this.transformTemplate],
            ['LINK', this.transformLink],
            ['STYLE', this.transformStyle],
            ['IMG', this.transformImg],
            ['AUDIO', this.transformMultimedia],
            ['SOURCE', this.transformMultimedia],
            ['VIDEO', this.transformMultimedia],
            ['IFRAME', this.transformIframe],
            ['BASE', this.setUrl]
        ]);
    }

    /**
     * Converts a relative URL to absolute using document's base URL
     * @param {string} url - The URL to convert
     * @returns {string} Absolute URL
     */
    absoluteUrl(url) {
        return this.resolver.resolve(url);
    }

    /**
     * Gets the complete document as text (doctype + HTML)
     * @returns {string} The complete document text
     */
    getDomAsText() {
        return `${this.doctype}${this.dom.outerHTML}`;
    }

    /**
     * Transforms the DOM by processing all nodes and downloading resources
     * @async
     */
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

    /**
     * Recursively walks and transforms the DOM tree
     * @async
     * @param {Node} node - The node to process
     * @returns {Promise} Promise that resolves when all nodes are processed
     */
    async walkDOM(node) {
        await this.transformNode(node);
        const children = [...node.childNodes];
        return Promise.allSettled(children.map(async (node) => {
            await this.walkDOM(node).catch(e => console.log("Error while transforming DOM:", e));
        }));
    }

    /**
     * Transforms a single node based on its type
     * @async
     * @param {Node} node - The node to transform
     */
    async transformNode(node) {
        if (node.nodeType !== Node.ELEMENT_NODE) {
            return;
        }
        this.sanitizer.sanitizeAttributes(node);
        await this.rewriteAttributes(node);
        const transformFunction = this.nodeTransformFunctions.get(node.nodeName);
        if (transformFunction) {
            try {
                await transformFunction.call(this, node);
            } catch(e) {
                console.log("Error in transformer function " + transformFunction.name + ":", e);
            }
        }
    }

    /**
     * Transforms LINK elements (stylesheets, icons, preloads)
     * @async
     * @param {HTMLLinkElement} node - The link element to transform
     */
    async transformLink(node) {
        const rel = (node.getAttribute('rel') || '').trim().toLowerCase();
        let res = null;
        switch (rel) {
            case 'stylesheet':
                if (!node.attributes.href) {
                    return;
                }
                const cssHref = this.absoluteUrl(node.attributes.href.nodeValue);
                res = await this.resources.create(cssHref);
                if (res) {
                    await res.updateContent(await this.sanitizer.sanitizeCSS(res.content, cssHref));
                    node.setAttribute('href', res.src);
                } else {
                    node.removeAttribute('href', '');
                }
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
                node.removeAttribute('href');
                break;
            case 'modulepreload':
                node.remove();
                return;
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
                        node.remove();
                        return;
                    case 'font':
                        res = await this.resources.create(this.absoluteUrl(href));
                        if (res) {
                            node.setAttribute('href', res.src);
                        } else {
                            node.removeAttribute('href');
                        }
                        break;
                    case 'image':
                    case 'style':
                        if(node.hasAttribute('imagesrcset')) {
                            node.removeAttribute('imagesrcset');
                        }
                        const cssHref = this.absoluteUrl(href);
                        res = await this.resources.create(cssHref);
                        if (res) {
                            await res.updateContent(await this.sanitizer.sanitizeCSS(res.content, cssHref));
                            node.setAttribute('href', res.src);
                        } else {
                            node.removeAttribute('href');
                        }
                        break;
                    case 'document':
                    case 'embed':
                    case 'image':
                    case 'audio':
                    case 'object':
                        // TODO handle preloading of the above types
                        node.removeAttribute('href');
                        break;

                }
                break;
        }
    }

    /**
     * Transforms STYLE elements by sanitizing CSS
     * @async
     * @param {HTMLStyleElement} node - The style element to transform
     */
    async transformStyle(node) {
        const innerText = await this.sanitizer.sanitizeCSS(node.textContent, this.absoluteUrl());
        node.textContent = innerText;
    }

    /**
     * Transforms IMG elements by downloading and embedding images
     * @async
     * @param {HTMLImageElement} node - The image element to transform
     */
    async transformImg(node) {
        if (node.getAttribute('src') && !node.getAttribute('src').startsWith('data:')) {
            const src = this.absoluteUrl(node.getAttribute('src'));
            const res = await this.resources.create(src);
            if (res) {
                node.setAttribute('src', res.src);
            } else {
                node.removeAttribute('src');
            }
        }
        if (node.getAttribute('srcset')) {
            // ignore srcest if we have src
            if (node.getAttribute('src')) {
                node.removeAttribute('srcset');
            } else {
                let val = node.getAttribute('srcset');
                let newParts = [];
                for (let s of val.split(',')) {
                    let srcParts = s.trim().split(' ');
                    const res = await this.resources.create(this.absoluteUrl(srcParts[0]));
                    if (res) {
                        srcParts[0] = res.src;
                        newParts.push(srcParts.join(' '));
                    }
                }
                node.setAttribute('srcset', newParts.join(', '));
            }
        }
    }

    /**
     * Transforms audio, video and source elements, rewrite references to absolute URLs so the omnom server can download them
     * @async
     * @param {HTMLImageElement} node - The audio/video/source element to transform
     */
    async transformMultimedia(node) {
        if (node.getAttribute('src') && !node.getAttribute('src').startsWith('data:')) {
            this.multimedia_count++;
            node.setAttribute('src', this.absoluteUrl(node.getAttribute('src')));
        }
        if (node.getAttribute('srcset')) {
            if (node.getAttribute('src')) {
                node.removeAttribute('srcset');
            } else {
                this.multimedia_count++;
                let newParts = [];
                for (let s of node.getAttribute('srcset').split(',')) {
                    let srcParts = s.trim().split(' ');
                    srcParts[0] = this.absoluteUrl(srcParts[0]);
                    newParts.push(srcParts.join(' '));
                }
                node.setAttribute('srcset', newParts.join(', '));
            }
        }
    }

    /**
     * Transforms IFRAME elements by embedding their content
     * @async
     * @param {HTMLIFrameElement} node - The iframe element to transform
     */
    async transformIframe(node) {
        const dataHtmlAttr = 'data-omnom-iframe-html';
        const dataUrlAttr = 'data-omnom-iframe-url';
        if (node.hasAttribute(dataHtmlAttr)) {
            let iframeUrl = node.getAttribute(dataUrlAttr);
            if (iframeUrl == "about:blank") {
                iframeUrl = this.absoluteUrl();
            }
            let iframeHtml = base64Decode(node.getAttribute(dataHtmlAttr));
            let iframe = new Document(iframeHtml, '', iframeUrl, '<!DOCTYPE html>', '', {});
            await iframe.transformDom();
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

    /**
     * Transforms TEMPLATE elements by processing their content
     * @async
     * @param {HTMLTemplateElement} node - The template element to transform
     */
    async transformTemplate(node) {
        await this.walkDOM(node.content);
    }

    /**
     * Processes BASE elements to update the base URL
     * @async
     * @param {HTMLBaseElement} node - The base element to process
     */
    async setUrl(node) {
        this.resolver.setBaseUrl(node.getAttribute('href'));
        node.removeAttribute('href');
    }

    /**
     * Rewrites node attributes (sanitize event handlers, resolve URLs, sanitize styles)
     * @async
     * @param {HTMLElement} node - The element whose attributes to rewrite
     * @returns {Promise} Promise that resolves when all attributes are processed
     */
    async rewriteAttributes(node) {
        const nodeAttributeArray = [...node.attributes];
        return Promise.allSettled(nodeAttributeArray.map(async (attr) => {
            if (attr.nodeName.startsWith('on') || attr.nodeValue.startsWith('javascript:')) {
                attr.nodeValue = '';
            }
            if (attr.nodeName == 'href' && node.nodeName != 'BASE') {
                attr.nodeValue = this.absoluteUrl(attr.nodeValue);
            }
            if (attr.nodeName == 'style') {
                const sanitizedValue = await this.sanitizer.sanitizeCSS(`a{${attr.nodeValue}}`, this.absoluteUrl());
                attr.nodeValue = sanitizedValue.substr(4, sanitizedValue.length - 6);
            }
        }));
    }
}

export { Document }
