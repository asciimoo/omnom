// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

import { absoluteURL } from './utils';
import { resources } from './resources';

const cssSanitizeFunctions = new Map([
    ['CSSStyleRule', sanitizeStyleRule],
    ['CSSImportRule', sanitizeImportRule],
    ['CSSMediaRule', sanitizeMediaRule],
    ['CSSFontFaceRule', sanitizeFontFaceRule],
    ['CSSPageRule', sanitizePageRule],
    ['CSSKeyframesRule', sanitizeKeyframesRule],
    ['CSSKeyframeRule', sanitizeKeyframeRule],
    ['CSSNamespaceRule', unknownRule], // XML only
    ['CSSCounterStyleRule', sanitizeCounterStyleRule],
    ['CSSSupportsRule', sanitizeSupportsRule],
    ['CSSDocumentRule', unknownRule], // FF only
    ['CSSFontFeatureValuesRule', unknownRule], // FF only
    ['CSSViewportRule', unknownRule], // IE only
]);

async function sanitizeStyleRule(rule, baseURL) {
    return await sanitizeCSSRule(rule, baseURL);
}

async function sanitizeImportRule(rule, baseURL) {
    // TODO handle import loops
    let href = absoluteURL(baseURL, rule.href);
    let res = await resources.create(href);
    await res.updateContent(await sanitizeCSS(res.content, href));
    return `@import url("${res.src}") ${rule.media};`;
}

async function sanitizeMediaRule(rule, baseURL) {
    const cssRuleArray = [...rule.cssRules];
    let cssResult = '';
    await Promise.allSettled(cssRuleArray.map(async (r, index) => {
        const css = await sanitizeCSSRule(r, baseURL);
        cssResult += css;
    }));
    return `@media ${rule.media.mediaText}{${cssResult}}`;
}

async function sanitizeFontFaceRule(rule, baseURL) {
    const fontRule = await sanitizeCSSFontFace(rule, baseURL);
    return fontRule ? fontRule : rule.cssText;
}

async function sanitizePageRule(rule, baseURL) {
    return rule.cssText;
}

async function sanitizeKeyframesRule(rule, baseURL) {
    let cssResult = await sanitizeCSS(rule.cssRules, baseURL);
    return `@keyframes ${rule.name}{${cssResult}}`;
}

async function sanitizeKeyframeRule(rule, baseURL) {
    return await sanitizeStyleRule(rule);
}

async function sanitizeSupportsRule(rule, baseURL) {
    let cssResult = await sanitizeCSS(rule.cssRules, baseURL);
    return `@supports ${rule.conditionText}{${cssResult}}`;
}

async function sanitizeCounterStyleRule(rule, baseURL) {
    return rule.cssText;
}

async function unknownRule(rule) {
    console.log('MEEEH, unknown css rule type: ', rule);
    return Promise.reject('MEEEH, unknown css rule type: ', rule);
}

async function sanitizeCSSRule(r, baseURL) {
    // huh? how can r be undefined?....
    if (!r || !r.style) {
        return '';
    }
    await sanitizeCSSBgImage(r, baseURL);
    await sanitizeCSSListStyleImage(r, baseURL);
    await sanitizeCSSContentImage(r, baseURL);
    await sanitizeCSSMask(r, baseURL);
    return r.cssText;
}

function parseCssUrls(rule) {
    let ret = new Set();
    for (let m of rule.matchAll(/url\(\"[^\"]+\"\)/g)) {
        ret.add(m[0].substring(5, m[0].length-2));
    }
    return ret;
}

async function sanitizeCSSBgImage(r, baseURL) {
    for (let u of parseCssUrls(r.style.backgroundImage)) {
        if (!u || u.startsWith('data:')) {
            continue;
        }
        const href = absoluteURL(baseURL, u);
        let res = await resources.create(href);
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
}

async function sanitizeCSSListStyleImage(r, baseURL) {
    const lsi = r.style.listStyleImage;
    if (lsi && lsi.startsWith('url("') && lsi.endsWith('")')) {
        const iURL = absoluteURL(baseURL, lsi.substring(5, lsi.length - 2));
        if (!iURL.startsWith('data:')) {
            let res = await resources.create(iURL);
            if (res) {
                try {
                    r.style.listStyleImage = `url('${res.src}')`;
                } catch (error) {
                    console.log('failed to set list-style-image:', error);
                    r.style.listStyleImage = '';
                }
            } else {
                r.style.listStyleImage = '';
            }
        }
    }
}

async function sanitizeCSSContentImage(r, baseURL) {
    const ci = r.style.content;
    if (ci && ci.startsWith('url("') && ci.endsWith('")')) {
        const bgURL = absoluteURL(baseURL, ci.substring(5, ci.length - 2));
        if (!bgURL.startsWith('data:')) {
            let res = await resources.create(bgURL);
            if (res) {
                try {
                    r.style.content = `url('${res.src}')`;
                } catch (error) {
                    console.log('failed to set content image: ', error);
                    r.style.content = '';
                }
            } else {
                console.log('failed to get content image: ', error);
                r.style.content = '';
            }
        }
    }
}

async function sanitizeCSSMask(r, baseURL) {
    await fixURL(r, "maskImage", baseURL);
}

async function fixURL(r, name, baseURL) {
    const attr = r.style[name];
    if (attr && attr.startsWith('url("') && attr.endsWith('")')) {
        const u = attr.substring(5, attr.length - 2);
        const bgURL = absoluteURL(baseURL, u);
        if (!bgURL.startsWith('data:')) {
            let res = await resources.create(bgURL);
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
}

async function sanitizeCSSFontFace(r, baseURL) {
    const src = r.style.getPropertyValue('src');
    const srcParts = src.split(/\s+/);
    let changed = false;
    for (const i in srcParts) {
        const part = srcParts[i];
        if (part && part.startsWith('url("') && part.endsWith('")')) {
            const iURL = absoluteURL(baseURL, part.substring(5, part.length - 2));
            if (!iURL.startsWith('data:')) {
                let res = await resources.create(iURL);
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
}

function parseCSS(styleContent) {
    const doc = document.implementation.createHTMLDocument('');
    const styleElement = document.createElement('style');

    styleElement.textContent = styleContent;
    // the style will only be parsed once it is added to a document
    doc.body.appendChild(styleElement);
    return styleElement.sheet.cssRules;
}

async function sanitizeCSS(rules, baseURL) {
    if (rules.constructor == ArrayBuffer || rules.constructor == Uint8Array) {
        let dec = new TextDecoder("utf-8");
        rules = dec.decode(rules);
    }
    if (typeof rules === 'string' || rules instanceof String) {
        rules = parseCSS(rules);
    }
    const cssMap = new Map();
    const rulesArray = [...rules];
    await Promise.allSettled(rulesArray.map(async (r, index) => {
        const sanitizeFunction = cssSanitizeFunctions.get(r.constructor.name);
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
}

function sanitizeAttributes(el) {
    let attrs = [...el.attributes];
    for(let i in attrs) {
        let key = attrs[i].nodeName;
        let val = attrs[i].nodeValue;
        if(key.toLowerCase().startsWith("on")) {
            el.removeAttribute(key);
        }
    }
}

export {
    sanitizeCSS,
    sanitizeAttributes,
};
