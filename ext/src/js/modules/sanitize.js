// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

/**
 * @fileoverview CSS and HTML sanitization utilities.
 * Sanitizes CSS rules, resolves URLs in stylesheets, and removes potentially harmful attributes.
 */

import { absoluteURL } from './utils';

/**
 * Sanitizes CSS and HTML content for safe snapshotting
 * @class
 */
class Sanitizer {
    /**
     * Creates a Sanitizer instance
     * @param {ResourceStorage} resources - Resource storage for managing downloaded assets
     */
    constructor(resources) {
        this.resources = resources;

        this.sanitizeStyleRule = async (rule, baseURL) => {
            return await this.sanitizeCSSRule(rule, baseURL);
        };
        this.sanitizeImportRule = async (rule, baseURL) => {
            // TODO handle import loops
            let href = absoluteURL(baseURL, rule.href);
            let res = await this.resources.create(href);
            await res.updateContent(await this.sanitizeCSS(res.content, href));
            return `@import url("${res.src}") ${rule.media};`;
        };
        this.sanitizeMediaRule = async (rule, baseURL) => {
            let cssResult = await this.sanitizeCSS(rule.cssRules, baseURL);
            return `@media ${rule.media.mediaText}{${cssResult}}`;
        };
        this.sanitizeFontFaceRule = async (rule, baseURL) => {
            const fontRule = await this.sanitizeCSSFontFace(rule, baseURL);
            return fontRule ? fontRule : rule.cssText;
        };
        this.sanitizePageRule = async (rule, baseURL) => {
            return rule.cssText;
        };
        this.sanitizeKeyframesRule = async (rule, baseURL) => {
            let cssResult = await this.sanitizeCSS(rule.cssRules, baseURL);
            return `@keyframes ${rule.name}{${cssResult}}`;
        };
        this.sanitizeKeyframeRule = async (rule, baseURL) => {
            return await this.sanitizeStyleRule(rule);
        };
        this.sanitizeSupportsRule = async (rule, baseURL) => {
            let cssResult = await this.sanitizeCSS(rule.cssRules, baseURL);
            return `@supports ${rule.conditionText}{${cssResult}}`;
        };
        this.sanitizeCounterStyleRule = async (rule, baseURL) => {
            return rule.cssText;
        };
        this.sanitizePropertyRule = async (rule, baseURL) => {
            // TODO check if initial-value requires sanitization
            return rule.cssText;
        };
        this.sanitizeViewTransitionRule = async (rule, baseURL) => {
            // TODO check if initial-value requires sanitization
            return rule.cssText;
        };
        this.sanitizeContainerRule = async (rule, baseURL) => {
            let cssResult = await this.sanitizeCSS(rule.cssRules, baseURL);
            return `@container ${rule.conditionText}{${cssResult}}`;
        };
        this.sanitizeLayerBlockRule = async (rule, baseURL) => {
            let name = '';
            if(rule.name) {
                name = rule.name;
            }
            if(rule.nameList) {
                name = rule.nameList.join(', ');
            }
            let cssResult = await this.sanitizeCSS(rule.cssRules, baseURL);
            if(cssResult) {
                return `@layer ${name}{${cssResult}}`;
            }
            return `@layer ${name}`;
        };
        this.sanitizeLayerStatementRule = async (rule, baseURL) => {
            return rule.cssText;
        }
        this.unknownRule = async (rule) => {
            console.log('MEEEH, unknown css rule type: ', rule);
            return Promise.reject('MEEEH, unknown css rule type: ', rule);
        };
        this.sanitizeCSSRule = async (r, baseURL) => {
            // huh? how can r be undefined?....
            if (!r || !r.style) {
                return '';
            }
            for(let prop of r.style) {
                if(prop.startsWith('--')) {
                    await this.fixURL(r, prop, baseURL);
                    continue
                }
                switch(prop) {
                    case 'background-image':
                    case 'list-style-image':
                    case 'content':
                    case 'mask-image':
                        await this.fixURL(r, prop, baseURL);
                        break;
                }

            }
            return r.cssText;
        };
        this.parseCssUrls = (rule) => {
            let ret = new Set();
            for (let m of rule.matchAll(/url\(([\"\']?)([^\)\"\']+)\1\)/g)) {
                ret.add(m[2]);
            }
            return ret;
        };
        this.fixURL = async (r, name, baseURL) => {
            const attr = r.style.getPropertyValue(name);
            if(!attr) {
                return;
            }
            for (let u of this.parseCssUrls(attr)) {
                if (!u || u.startsWith('data:')) {
                    continue;
                }
                const href = absoluteURL(baseURL, u);
                let res = await this.resources.create(href);
                if (res) {
                    try {
                        r.style.setProperty(name, r.style.getPropertyValue(name).replaceAll(u, res.src));
                    } catch (error) {
                        r.style.setProperty(name, '');
                    }
                } else {
                    r.style.setProperty(name, '');
                }
            }
        };
        this.sanitizeCSSFontFace = async (r, baseURL) => {
            const src = r.style.getPropertyValue('src');
            const srcParts = src.split(/\s+/);
            let changed = false;
            for (const i in srcParts) {
                const part = srcParts[i];
                if (part && part.startsWith('url("') && part.endsWith('")')) {
                    const iURL = absoluteURL(baseURL, part.substring(5, part.length - 2));
                    if (!iURL.startsWith('data:')) {
                        let res = await this.resources.create(iURL);
                        if (res) {
                            srcParts[i] = `url('${res.src}')`;
                        } else {
                            srcParts[i] = '';
                        }
                        changed = true;
                    }
                }
            }
            if (changed) {
                try {
                    return `@font-face {${r.style.cssText.replace(src, srcParts.join(' '))}}`;
                } catch (error) {
                    console.log('failed to set font-src:', error);
                    r.style.src = '';
                }
            }
            return '';
        };
        this.parseCSS = (styleContent) => {
            const doc = document.implementation.createHTMLDocument('');
            const styleElement = document.createElement('style');

            styleElement.textContent = styleContent;
            // the style will only be parsed once it is added to a document
            doc.body.appendChild(styleElement);
            return styleElement.sheet.cssRules;
        };
        this.sanitizeCSS = async (rules, baseURL) => {
            if (rules.constructor == ArrayBuffer || rules.constructor == Uint8Array) {
                let dec = new TextDecoder("utf-8");
                rules = dec.decode(rules);
            }
            if (typeof rules === 'string' || rules instanceof String) {
                rules = this.parseCSS(rules);
            }
            const cssMap = new Map();
            const rulesArray = [...rules];
            await Promise.allSettled(rulesArray.map(async (r, index) => {
                const sanitizeFunction = this.cssSanitizeFunctions.get(r.constructor.name);
                if (sanitizeFunction) {
                    const css = await sanitizeFunction(r, baseURL).catch(err => console.log(err));
                    cssMap.set(index, css);
                } else {
                    this.unknownRule(r, baseURL);
                }
            }));
            const sortedCss = new Map([...cssMap.entries()].sort((e1, e2) => e1[0] - e2[0]));
            const result = [...sortedCss.values()].join('');
            return result;
        };
        this.sanitizeAttributes = (el) => {
            let attrs = [...el.attributes];
            for(let i in attrs) {
                let key = attrs[i].nodeName;
                let val = attrs[i].nodeValue;
                if(key.toLowerCase().startsWith("on")) {
                    el.removeAttribute(key);
                }
            }
        };
        this.cssSanitizeFunctions = new Map([
            ['CSSStyleRule', this.sanitizeStyleRule],
            ['CSSImportRule', this.sanitizeImportRule],
            ['CSSMediaRule', this.sanitizeMediaRule],
            ['CSSFontFaceRule', this.sanitizeFontFaceRule],
            ['CSSPageRule', this.sanitizePageRule],
            ['CSSKeyframesRule', this.sanitizeKeyframesRule],
            ['CSSKeyframeRule', this.sanitizeKeyframeRule],
            ['CSSContainerRule', this.sanitizeContainerRule],
            ['CSSLayerBlockRule', this.sanitizeLayerBlockRule],
            ['CSSLayerStatementRule', this.sanitizeLayerStatementRule],
            ['CSSPropertyRule', this.sanitizePropertyRule],
            ['CSSViewTransitionRule', this.sanitizeViewTransitionRule],
            ['CSSNamespaceRule', this.unknownRule], // XML only
            ['CSSCounterStyleRule', this.sanitizeCounterStyleRule],
            ['CSSSupportsRule', this.sanitizeSupportsRule],
            ['CSSDocumentRule', this.unknownRule], // FF only
            ['CSSFontFeatureValuesRule', this.unknownRule], // FF only
            ['CSSViewportRule', this.unknownRule], // IE only
        ]);
    }
}



export {
    Sanitizer
};
