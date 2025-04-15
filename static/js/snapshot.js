!function(t,e){"object"==typeof exports&&"object"==typeof module?module.exports=e():"function"==typeof define&&define.amd?define("webapp_snapshot",[],e):"object"==typeof exports?exports.webapp_snapshot=e():t.webapp_snapshot=e()}(self,(()=>(()=>{"use strict";var t={d:(e,s)=>{for(var r in s)t.o(s,r)&&!t.o(e,r)&&Object.defineProperty(e,r,{enumerable:!0,get:s[r]})},o:(t,e)=>Object.prototype.hasOwnProperty.call(t,e),r:t=>{"undefined"!=typeof Symbol&&Symbol.toStringTag&&Object.defineProperty(t,Symbol.toStringTag,{value:"Module"}),Object.defineProperty(t,"__esModule",{value:!0})}},e={};t.r(e),t.d(e,{createOmnomSnapshot:()=>C});const s=chrome;let r="",n=!1;function i(t){return t.ok?Promise.resolve(t):Promise.reject(t)}function o(t,e){return new URL(e,t).href}async function a(t){if(t)return void(r=t);const e=await new Promise((t=>{s.tabs.query({active:!0,currentWindow:!0},(([e])=>t(e)))}));e&&(r=e.url)}function c(){return n}class l{constructor(t){this.url=t,this.hasBaseUrl=!1}resolve(t){return t?t.startsWith("data:")?t:this.hasBaseUrl&&!t.startsWith("/")&&-1==t.search(/^[a-zA-Z]+:\/\//)?this.url+t:new URL(t,this.url).href:this.url}setBaseUrl(t){this.hasBaseUrl=!0,this.url=this.resolve(t)}}function u(t){return btoa(unescape(encodeURIComponent(t)))}const h={DOWNLOADING:"downloading",DOWNLOADED:"downloaded",FAILED:"failed"};let m=0,f=0,d=0,y=null;async function p(t){if(!t||(t||"").startsWith("data:"))return t;t=function(t){return new URL(t,r).href}(t),console.log("fetching ",t);const e={method:"GET",cache:c?"no-cache":"default"},s=new Request(t,e);w(h.DOWNLOADING);let n=!1;const o=await fetch(s,e).then(i).catch((()=>{w(h.FAILED),n=!0}));if(n)return"";const a=o.headers.get("Content-Type");if(w(h.DOWNLOADED),a&&-1!=a.toLowerCase().search("text"))return await o.text();return`${`data:${a};base64,`}${function(t){let e="";return[].slice.call(new Uint8Array(t)).forEach((t=>e+=String.fromCharCode(t))),btoa(e)}(await o.arrayBuffer())}`}function w(t){switch(t){case h.DOWNLOADING:f++;break;case h.DOWNLOADED:m++;break;case h.FAILED:d++}null!==y&&y.next({downloadCount:f,downloadedCount:m,failedCount:d})}const S=new Map([["jpeg","jpg"]]);class b{constructor(t,e,s){this.content=t,this.mimetype=e,this.filename=s,this.extension="unknown",e&&(this.extension=e.split(" ")[0].split("/").pop().toLowerCase().split("+")[0].split(";")[0]),S.has(this.extension)&&(this.extension=S.get(this.extension)),this.src=""}async sha(){this.sha256sum=await async function(t){"String"==t.__proto__.constructor.name&&(t=(new TextEncoder).encode(t));const e=await crypto.subtle.digest("SHA-256",t);return Array.from(new Uint8Array(e)).map((t=>t.toString(16).padStart(2,"0"))).join("")}(this.content),this.src=`../../resources/${this.sha256sum[0]}${this.sha256sum[1]}/${this.sha256sum}.${this.extension}`}async updateContent(t){this.content=t,await this.sha()}}class g{constructor(){this.resources=new Map([])}async create(t){if(this.resources.has(t))return this.resources.get(t);let e=await async function(t){const e={method:"GET",cache:c?"no-cache":"default"};w(h.DOWNLOADING);const s=new Request(t,e);let r=!1;const n=await fetch(s,e).then(i).catch((()=>{r=!0,w(h.FAILED)}));return r?"":(w(h.DOWNLOADED),n)}(t);if(!e)return;const s=await e.arrayBuffer();if(!s)return;const r=e.headers.get("Content-Type"),n=new URL(t).pathname.split("/").pop();let o=new b(s,r,n);return await o.sha(),this.resources.set(t,o),o}getAll(){return this.resources.values()}}class A{constructor(t){this.resources=t,this.sanitizeStyleRule=async(t,e)=>await this.sanitizeCSSRule(t,e),this.sanitizeImportRule=async(t,e)=>{let s=o(e,t.href),r=await this.resources.create(s);return await r.updateContent(await this.sanitizeCSS(r.content,s)),`@import url("${r.src}") ${t.media};`},this.sanitizeMediaRule=async(t,e)=>{let s=await this.sanitizeCSS(t.cssRules,e);return`@media ${t.media.mediaText}{${s}}`},this.sanitizeFontFaceRule=async(t,e)=>{const s=await this.sanitizeCSSFontFace(t,e);return s||t.cssText},this.sanitizePageRule=async(t,e)=>t.cssText,this.sanitizeKeyframesRule=async(t,e)=>{let s=await this.sanitizeCSS(t.cssRules,e);return`@keyframes ${t.name}{${s}}`},this.sanitizeKeyframeRule=async(t,e)=>await this.sanitizeStyleRule(t),this.sanitizeSupportsRule=async(t,e)=>{let s=await this.sanitizeCSS(t.cssRules,e);return`@supports ${t.conditionText}{${s}}`},this.sanitizeCounterStyleRule=async(t,e)=>t.cssText,this.sanitizePropertyRule=async(t,e)=>t.cssText,this.sanitizeContainerRule=async(t,e)=>{let s=await this.sanitizeCSS(t.cssRules,e);return`@container ${t.conditionText}{${s}}`},this.sanitizeLayerBlockRule=async(t,e)=>{let s="";t.name&&(s=t.name),t.nameList&&(s=t.nameList.join(", "));let r=await this.sanitizeCSS(t.cssRules,e);return r?`@layer ${s}{${r}}`:`@layer ${s}`},this.sanitizeLayerStatementRule=async(t,e)=>t.cssText,this.unknownRule=async t=>(console.log("MEEEH, unknown css rule type: ",t),Promise.reject("MEEEH, unknown css rule type: ",t)),this.sanitizeCSSRule=async(t,e)=>{if(!t||!t.style)return"";for(let s of t.style)if(s.startsWith("--"))console.log("YOOOO",s,t.style.getPropertyValue(s)),await this.fixURL(t,s,e),console.log("YOOOOHOHOOO",s,t.style.getPropertyValue(s));else switch(s){case"background-image":case"list-style-image":case"content":case"mask-image":await this.fixURL(t,s,e)}return t.cssText},this.parseCssUrls=t=>{let e=new Set;for(let s of t.matchAll(/url\(([\"\']?)([^\)\"\']+)\1\)/g))e.add(s[2]);return e},this.fixURL=async(t,e,s)=>{const r=t.style.getPropertyValue(e);if(r)for(let n of this.parseCssUrls(r)){if(!n||n.startsWith("data:"))continue;const r=o(s,n);console.log("trying to set URL for",e,n,r);let i=await this.resources.create(r);if(i){console.log("setting URL for",e,n,i.src);try{t.style.setProperty(e,t.style.getPropertyValue(e).replaceAll(n,i.src)),console.log("url set")}catch(s){console.log(`failed to set ${e} css url property: `,s),t.style.setProperty(e,"")}}else console.log(`failed to set ${e} css url property: cannot download resource`),t.style.setProperty(e,"")}},this.sanitizeCSSFontFace=async(t,e)=>{const s=t.style.getPropertyValue("src"),r=s.split(/\s+/);let n=!1;for(const t in r){const s=r[t];if(s&&s.startsWith('url("')&&s.endsWith('")')){const i=o(e,s.substring(5,s.length-2));if(!i.startsWith("data:")){let e=await this.resources.create(i);r[t]=e?`url('${e.src}')`:"",n=!0}}}if(n)try{return`@font-face {${t.style.cssText.replace(s,r.join(" "))}}`}catch(e){console.log("failed to set font-src:",e),t.style.src=""}return""},this.parseCSS=t=>{const e=document.implementation.createHTMLDocument(""),s=document.createElement("style");return s.textContent=t,e.body.appendChild(s),s.sheet.cssRules},this.sanitizeCSS=async(t,e)=>{if(t.constructor==ArrayBuffer||t.constructor==Uint8Array){t=new TextDecoder("utf-8").decode(t)}("string"==typeof t||t instanceof String)&&(t=this.parseCSS(t));const s=new Map,r=[...t];await Promise.allSettled(r.map((async(t,r)=>{const n=this.cssSanitizeFunctions.get(t.constructor.name);if(n){const i=await n(t,e).catch((t=>console.log(t)));s.set(r,i)}else this.unknownRule(t,e)})));return[...new Map([...s.entries()].sort(((t,e)=>t[0]-e[0]))).values()].join("")},this.sanitizeAttributes=t=>{let e=[...t.attributes];for(let s in e){let r=e[s].nodeName;e[s].nodeValue;r.toLowerCase().startsWith("on")&&t.removeAttribute(r)}},this.cssSanitizeFunctions=new Map([["CSSStyleRule",this.sanitizeStyleRule],["CSSImportRule",this.sanitizeImportRule],["CSSMediaRule",this.sanitizeMediaRule],["CSSFontFaceRule",this.sanitizeFontFaceRule],["CSSPageRule",this.sanitizePageRule],["CSSKeyframesRule",this.sanitizeKeyframesRule],["CSSKeyframeRule",this.sanitizeKeyframeRule],["CSSContainerRule",this.sanitizeContainerRule],["CSSLayerBlockRule",this.sanitizeLayerBlockRule],["CSSLayerStatementRule",this.sanitizeLayerStatementRule],["CSSPropertyRule",this.sanitizePropertyRule],["CSSNamespaceRule",this.unknownRule],["CSSCounterStyleRule",this.sanitizeCounterStyleRule],["CSSSupportsRule",this.sanitizeSupportsRule],["CSSDocumentRule",this.unknownRule],["CSSFontFeatureValuesRule",this.unknownRule],["CSSViewportRule",this.unknownRule]])}}class R{constructor(t,e,s,r,n,i){this.doctype=r,this.dom=document.createElement("html"),this.iframes=[],this.favicon=null,this.dom.innerHTML=t,this.originalLength=t.length,this.resolver=new l(s),this.resources=new g,this.sanitizer=new A(this.resources),this.text=e;for(const t in i)this.dom.setAttribute(t,i[t]);this.nodeTransformFunctions=new Map([["SCRIPT",t=>t.remove()],["LINK",this.transformLink],["STYLE",this.transformStyle],["IMG",this.transformImg],["IFRAME",this.transformIframe],["BASE",this.setUrl]])}absoluteUrl(t){return this.resolver.resolve(t)}getDomAsText(){return`${this.doctype}${this.dom.outerHTML}`}async transformDom(){if(await this.walkDOM(this.dom),!this.favicon&&(this.favicon=await p(this.absoluteUrl("/favicon.ico")),this.favicon)){const t=document.createElement("link");t.setAttribute("rel","icon"),t.setAttribute("href",this.favicon),this.dom.getElementsByTagName("head")[0].appendChild(t)}}async walkDOM(t){await this.transformNode(t);const e=[...t.childNodes];return Promise.allSettled(e.map((async t=>{await this.walkDOM(t).catch((t=>console.log("Error while transforming DOM:",t)))})))}async transformNode(t){if(t.nodeType!==Node.ELEMENT_NODE)return;this.sanitizer.sanitizeAttributes(t),await this.rewriteAttributes(t);const e=this.nodeTransformFunctions.get(t.nodeName);if(e)try{await e.call(this,t)}catch(t){console.log("Error in transformer function "+e.name+":",t)}}async transformLink(t){let e=null;switch((t.getAttribute("rel")||"").trim().toLowerCase()){case"stylesheet":if(!t.attributes.href)return;const s=this.absoluteUrl(t.attributes.href.nodeValue);e=await this.resources.create(s),e?(await e.updateContent(await this.sanitizer.sanitizeCSS(e.content,s)),t.setAttribute("href",e.src)):t.removeAttribute("href","");break;case"icon":case"shortcut icon":case"apple-touch-icon":case"apple-touch-icon-precomposed":case"fluid-icon":const r=await p(this.absoluteUrl(t.getAttribute("href")));t.setAttribute("href",r),this.favicon||(this.favicon=r);break;case"preconnect":case"dns-prefetch":t.removeAttribute("href");break;case"preload":const n=t.getAttribute("href");if(!n)break;switch((t.getAttribute("as")||"").toLowerCase()){case"script":case"fetch":case"track":case"worker":case"document":case"embed":case"image":case"audio":case"object":t.removeAttribute("href");break;case"font":e=await this.resources.create(this.absoluteUrl(n)),e?t.setAttribute("href",e.src):t.removeAttribute("href");break;case"style":const s=this.absoluteUrl(n);e=await this.resources.create(s),e?(await e.updateContent(await this.sanitizer.sanitizeCSS(e.content,s)),t.setAttribute("href",e.src)):t.removeAttribute("href")}}}async transformStyle(t){const e=await this.sanitizer.sanitizeCSS(t.textContent,this.absoluteUrl());t.textContent=e}async transformImg(t){if(t.getAttribute("src")&&!t.getAttribute("src").startsWith("data:")){const e=this.absoluteUrl(t.getAttribute("src")),s=await this.resources.create(e);s?t.setAttribute("src",s.src):t.removeAttribute("src")}if(t.getAttribute("srcset"))if(t.getAttribute("src"))t.removeAttribute("srcset");else{let e=t.getAttribute("srcset"),s=[];for(let t of e.split(",")){let e=t.trim().split(" ");const r=await this.resources.create(this.absoluteUrl(e[0]));r&&(e[0]=r.src,s.push(e.join(" ")))}t.setAttribute("srcset",s.join(", "))}}async transformIframe(t){const e="data-omnom-iframe-html",s="data-omnom-iframe-url";if(t.hasAttribute(e)){let n=t.getAttribute(s);"about:blank"==n&&(n=this.absoluteUrl());let i=(r=t.getAttribute(e),decodeURIComponent(escape(atob(r)))),o=new R(i,"",n,"<!DOCTYPE html>","",{});await o.transformDom();const a=`data:text/html;base64,${u(o.getDomAsText())}`;return t.setAttribute("src",a),t.removeAttribute(e),void t.removeAttribute(s)}var r;if(!t.getAttribute("src"))return;const n=this.absoluteUrl(t.getAttribute("src"));for(let e of this.iframes)if(e.absoluteUrl()==n){await e.transformDom();const s=`data:text/html;base64,${u(e.getDomAsText())}`;return void t.setAttribute("src",s)}console.log("Meh, iframe not found: ",n),t.setAttribute("src","")}async setUrl(t){this.resolver.setBaseUrl(t.getAttribute("href")),t.removeAttribute("href")}async rewriteAttributes(t){const e=[...t.attributes];return Promise.allSettled(e.map((async e=>{if((e.nodeName.startsWith("on")||e.nodeValue.startsWith("javascript:"))&&(e.nodeValue=""),"href"==e.nodeName&&"BASE"!=t.nodeName&&(e.nodeValue=this.absoluteUrl(e.nodeValue)),"style"==e.nodeName){const t=await this.sanitizer.sanitizeCSS(`a{${e.nodeValue}}`,this.absoluteUrl());e.nodeValue=t.substr(4,t.length-6)}})))}}async function C(){const t=function(){const t=["ARTICLE","ASIDE","BLOCKQUOTE","DIV","DL","DT","FIGURE","FOOTER","H1","H2","H3","H4","H5","H6","LI","MAIN","NAV","P","SECTION","TD","TH"];function e(t){if(t.nodeType!=Node.ELEMENT_NODE)return NodeFilter.FILTER_ACCEPT;const e=window.getComputedStyle(t),s=t.getBoundingClientRect();return s.width<5||s.height<5||"none"==e.display||"hidden"==e.visibility||"0"==e.opacity?NodeFilter.FILTER_REJECT:NodeFilter.FILTER_ACCEPT}const s=function(t){let e,s=document.createNodeIterator(t,NodeFilter.SHOW_ELEMENT),r=0,n=[];for(;e=s.nextNode();){if(!e.shadowRoot)continue;let t=[];for(let s of e.shadowRoot.children)t.push(s.outerHTML);e.setAttribute("omnomshadowroot",r++);let s=document.createElement("template");s.innerHTML=t.join(""),n.push(s)}return n}(document.getRootNode()),r=document.documentElement,n=r.querySelectorAll("style");if(n)for(let t of n){let e;try{e=t.sheet?.cssRules}catch(t){console.log("failed to access to css, probably it comes from another extension: "+t);continue}if(e){const s=[...e].reduce(((t,e)=>t.concat(e.cssText)),"");t.textContent=s}}const i={html:r.cloneNode(!0),attributes:{},title:"",doctype:"",iframeCount:r.querySelectorAll("iframe").length,url:document.URL};var o,a;o=i.html,a=s,o.querySelectorAll("[omnomshadowroot]").forEach((t=>{let e=Number(t.getAttribute("omnomshadowroot"));t.prepend(a[e])})),i.text=function(s){const r=document.createTreeWalker(s,NodeFilter.SHOW_TEXT|NodeFilter.SHOW_ELEMENT,e);let n=[],i=(s.tagName,[]);for(;r.nextNode();){let e=r.currentNode;if(e.nodeType==Node.ELEMENT_NODE){if(t.includes(e.tagName)&&n.length>0){let t=n.join("").replace(/\s+/g," ").trim();t&&i.push(t),n=[]}}else e.nodeType==Node.TEXT_NODE&&n.push(e.nodeValue)}if(n.length>0){let t=n.join("").replace(/\s+/g," ").trim();t&&i.push(t)}return i}(document.body).join("|||"),document.doctype&&(i.doctype=(new XMLSerializer).serializeToString(document.doctype)),document.getElementsByTagName("title").length>0&&(i.title=document.getElementsByTagName("title")[0].innerText),[...r.attributes].forEach((t=>i.attributes[t.nodeName]=t.nodeValue));let c=r.querySelectorAll("canvas");if(c){let t=[];for(let e of c){let s=document.createElement("img");s.src=e.toDataURL(),t.push(s)}let e=i.html.querySelectorAll("canvas");for(let s in t)e[s].replaceWith(t[s])}let l=r.querySelectorAll("iframe");if(l){let t=[];for(let e of l){const s=e.contentDocument;s?t.push({html:btoa(unescape(encodeURIComponent(s.documentElement.outerHTML))),url:s.URL||document.URL}):t.push(0)}let e=i.html.querySelectorAll("iframe");for(let s in t)t[s]&&(e[s].setAttribute("data-omnom-iframe-html",t[s].html),e[s].setAttribute("data-omnom-iframe-url",t[s].url))}return i.html=i.html.outerHTML,i}(),e=new R(t.html,t.text,t.url,t.doctype,t.title,t.attributes);await a(t.url);const s=await async function(t){return await t.transformDom(),{dom:t.getDomAsText(),favicon:t.favicon}}(e),r={dom:s.dom,favicon:s.favicon,resources:[],text:t.text,title:t.title};for(let t of e.resources.getAll())t.content=Array.from(new Uint8Array(t.content)),r.resources.push(t);return r}return e})()));
//# sourceMappingURL=snapshot.js.map