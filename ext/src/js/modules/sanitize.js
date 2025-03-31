// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

import { absoluteURL } from './utils';

class Sanitizer {
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
            const cssRuleArray = [...rule.cssRules];
            let cssResult = '';
            await Promise.allSettled(cssRuleArray.map(async (r, index) => {
                const css = await this.sanitizeCSSRule(r, baseURL);
                cssResult += css;
            }));
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
        this.unknownRule = async (rule) => {
            console.log('MEEEH, unknown css rule type: ', rule);
            return Promise.reject('MEEEH, unknown css rule type: ', rule);
        };
        this.sanitizeCSSRule = async (r, baseURL) => {
            // huh? how can r be undefined?....
            if (!r || !r.style) {
                return '';
            }
            await this.sanitizeCSSBgImage(r, baseURL);
            await this.sanitizeCSSListStyleImage(r, baseURL);
            await this.sanitizeCSSContentImage(r, baseURL);
            await this.sanitizeCSSMask(r, baseURL);
            return r.cssText;
        };
        this.parseCssUrls = (rule) => {
            let ret = new Set();
            for (let m of rule.matchAll(/url\(\"[^\"]+\"\)/g)) {
                ret.add(m[0].substring(5, m[0].length-2));
            }
            return ret;
        };
        this.sanitizeCSSBgImage = async (r, baseURL) => {
            for (let u of this.parseCssUrls(r.style.backgroundImage)) {
                if (!u || u.startsWith('data:')) {
                    continue;
                }
                const href = absoluteURL(baseURL, u);
                let res = await this.resources.create(href);
                if (res) {
                    try {
                        r.style.backgroundImage = r.style.backgroundImage.replaceAll(u, res.src);
                    } catch (error) {
                        console.log('failed to set background image: ', error);
                        r.style.backgroundImage = '';
                        break;
                    }
                } else {
                    console.log('failed to download background image: ', u);
                    r.style.backgroundImage = '';
                    break;
                }
            }
        };
        this.sanitizeCSSListStyleImage = async (r, baseURL) => {
            await this.fixURL(r, "listStyleImage", baseURL);
        };

        this.sanitizeCSSContentImage = async (r, baseURL) => {
            await this.fixURL(r, "content", baseURL);
        };

        this.sanitizeCSSMask = async (r, baseURL) => {
            await this.fixURL(r, "maskImage", baseURL);
        };
        this.fixURL = async (r, name, baseURL) => {
            const attr = r.style[name];
            if (attr && attr.startsWith('url("') && attr.endsWith('")')) {
                const u = attr.substring(5, attr.length - 2);
                const iURL = absoluteURL(baseURL, u);
                if (!iURL.startsWith('data:')) {
                    let res = await this.resources.create(iURL);
                    if (res) {
                        try {
                            r.style[name] = `url('${res.src}')`;
                        } catch (error) {
                            console.log(`failed to set ${name} css url property: `, error);
                            r.style[name] = '';
                        }
                    } else {
                        console.log(`failed to set ${name} css url property: `, error);
                        r.style[name] = '';
                    }
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
                    unknownRule(r, baseURL);
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
