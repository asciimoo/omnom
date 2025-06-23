{{ define "full-content" }}
<div class="iframe-diff-header mb-3">
    <h2 class="title">Side by side snapshot diff of {{ Truncate .SURL 250 }}</h2>
    <div class="container is-fluid content">
        <a href="{{ URLFor "Snapshot diff side by side" }}?s1={{ .S2.Key }}&s2={{ .S1.Key }}">Swap sides</a><br />
        <a href="{{ URLFor "Snapshot diff" }}?s1={{ .S1.Key }}&s2={{ .S2.Key }}">Show differences</a>
    </div>
    <div class="columns mb-2">
        <div class="column">
            <a href="{{ URLFor "Snapshot" }}?sid={{ .S1.Key }}&bid={{ .S1.BookmarkID }}" class="is-size-3 pl-5">{{ .S1.CreatedAt | ToDateTime }}</a>
        </div>
        <div class="column">
            <a href="{{ URLFor "Snapshot" }}?sid={{ .S2.Key }}&bid={{ .S2.BookmarkID }}" class="is-size-3 pl-5">{{ .S2.CreatedAt | ToDateTime }}</a>
        </div>
    </div>
</div>
<div class="iframe-diff-wrapper">
    <noscript>{{ block "warning" KVData "Warning" "this feature requires javascript" "Tr" .Tr }}{{ end }}</noscript>
    <div class="columns">
        <div class="column">
            <iframe id="sn1" title="snapshot 1" scrolling="no"></iframe>
        </div>
        <div class="column">
            <iframe title="snapshot 2" id="sn2" scrolling="no"></iframe>
        </div>
    </div>
</div>
<script>
 let s1 = document.getElementById('sn1');
 let s2 = document.getElementById('sn2');
 let s1Loaded = false;
 let s2Loaded = false;
 function getMaxHeight(win) {
     let body = win.document.body;
     let html = win.document.documentElement;
     let bodyH = parseInt(win.getComputedStyle(body).height);
     return Math.max(body.scrollHeight, body.offsetHeight, html.clientHeight, html.scrollHeight, html.offsetHeight, bodyH)+"px";
 }
 function resizeIFrameToFitContent(iFrame) {
     let h = getMaxHeight(iFrame.contentWindow);
     iFrame.style.height = h;
     iFrame.parentNode.style.height = h;
 }
 function highlightDiffs(e1, e2) {
     if(e1.nodeName == 'SCRIPT' || e1.nodeName == 'STYLE') {
         return false;
     }
     if(Array(...e1.childNodes).every(n => n.nodeType != Node.ELEMENT_NODE)) {
         highlightElement(e1);
         return false;
     }
     let matchCount = 0;
     let nodeCount = 0;
     for(let c of e1.childNodes) {
         if(c.nodeType == Node.TEXT_NODE) {
             if(c.textContent.trim() && e2.textContent.indexOf(c.textContent) == -1) {
                let h = document.createElement("span");
                h.textContent = c.textContent;
                c.replaceWith(h);
                highlightElement(h);
             }
             continue;
         }
         if(c.nodeType != Node.ELEMENT_NODE) {
             continue;
         }
         nodeCount += 1;
         let i = document.createNodeIterator(e2, NodeFilter.SHOW_ELEMENT)
         let match = false;
         let n = undefined;
         while((n = i.nextNode())) {
             if(c.isEqualNode(n)) {
                 match = true;
                 break;
             }
         }
         if(match) {
             matchCount += 1;
             continue;
         }
         if(highlightDiffs(c, e2)) {
             highlightElement(c);
         }
     }
     return matchCount > 0 && matchCount == nodeCount;
 }
 function highlightElement(el) {
    el.style.backgroundColor = "#ffeaa7";
    el.style.color = "black";
    el.style.animation = "omnom-highlight 2s infinite";
 }

 function initIframe(el) {
     let p = new Promise((resolve, reject) => {
        el.addEventListener('load', function(e) {
            resizeIFrameToFitContent(el);
            resolve();
        });
     });
     return p;
 }

 let s1URL = "{{ SnapshotURL .S1.Key }}";
 let s2URL = "{{ SnapshotURL .S2.Key }}";
 let s1p = initIframe(s1);
 let s2p = initIframe(s2);
 function loadSnapshot(url, el) {
    return fetch(url).then(r => {
        return r.text();
    }).then(html => {
        const css = document.createElement("style");
        css.textContent = `@keyframes omnom-highlight {
            0% {background-color: #ffeaa7;}
            50% {background-color: #fdcb6e;}
            100% {background-color: #ffeaa7;}
        }`;
        const parser = new DOMParser();
        const doc = parser.parseFromString(html, "text/html");
        const base = document.createElement("base")
        base.setAttribute("href", "./static/data/snapshots/aa/");
        doc.querySelector("head").prepend(base);
        doc.querySelector("head").appendChild(css);
        el.srcdoc = doc.documentElement.outerHTML;
    }).catch(err => console.log(err));
 }
 Promise.all([loadSnapshot(s1URL, s1), loadSnapshot(s2URL, s2), s1p, s2p]).then(() => {
     highlightDiffs(s2.contentWindow.document.body, s1.contentWindow.document.body);
     highlightDiffs(s1.contentWindow.document.body, s2.contentWindow.document.body);
 });
</script>
{{ end }}
