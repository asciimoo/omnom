{{ define "full-content" }}
<div class="iframe-diff-wrapper">
    <h2 class="title">Side by side snapshot diff of {{ Truncate .SURL 250 }}</h2>
    <div class="container is-fluid content"><a href="{{ URLFor "Snapshot diff side by side" }}?s1={{ .S2.Key }}&s2={{ .S1.Key }}">swap sides</a></div>
    <noscript>{{ block "warning" KVData "Warning" "this feature requires javascript" "Tr" .Tr }}{{ end }}</noscript>
    <div class="columns">
        <div class="column">
            <a href="{{ URLFor "Snapshot" }}?sid={{ .S1.Key }}&bid={{ .S1.BookmarkID }}" class="is-size-3">{{ .S1.CreatedAt | ToDateTime }}</a>
            <iframe id="sn1" title="snapshot 1" scrolling="no" class="snapshot-iframe"></iframe>
        </div>
        <div class="column">
            <a href="{{ URLFor "Snapshot" }}?sid={{ .S2.Key }}&bid={{ .S2.BookmarkID }}" class="is-size-3">{{ .S2.CreatedAt | ToDateTime }}</a>
            <iframe title="snapshot 2" id="sn2" class="snapshot-iframe" scrolling="no"></iframe>
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
     if(Array(...e1.childNodes).every(n => n.nodeType != Node.ELEMENT_NODE)) {
         e1.style.backgroundColor = "#ffeaa7";
         e1.style.color = "black";
         e1.style.animation = "omnom-highlight 2s infinite";
         return false;
     }
     let matchCount = 0;
     let nodeCount = 0;
     for(let c of e1.childNodes) {
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
             c.style.backgroundColor = "#ffeaa7";
             c.style.color = "black";
             c.style.animation = "omnom-highlight 2s infinite";
         }
     }
     return matchCount > 0 && matchCount == nodeCount;
 }

 s2.addEventListener('load', function(e) {
     resizeIFrameToFitContent(s2);
     s2Loaded = true;
     if(s1Loaded) {
        highlightDiffs(s2.contentWindow.document.body, s1.contentWindow.document.body);
        s2Loaded = false;
     }
 });
 s1.addEventListener('load', function(e) {
     resizeIFrameToFitContent(s1);
     s1Loaded = true;
     if(s2Loaded) {
        highlightDiffs(s2.contentWindow.document.body, s1.contentWindow.document.body);
        s1Loaded = false;
     }
 });

 let s1URL = "{{ SnapshotURL .S1.Key }}";
 let s2URL = "{{ SnapshotURL .S2.Key }}";
 function loadSnapshot(url, el) {
    return fetch(url).then(r => {
        return r.text();
    }).then(html => {
        let css = `<style>@keyframes omnom-highlight {
            0% {background-color: #ffeaa7;}
            50% {background-color: #fdcb6e;}
            100% {background-color: #ffeaa7;}
        }</style>`;
        el.srcdoc = html.replace("<head>", `<head><base href="./static/data/snapshots/aa/">`+css);
        //d.open();
        //d.write();
        //d.close();
        //s1.setAttribute("src", "data:text/html;base64,"+btoa(html.replace("<head>", `<head><base href="./static/data/snapshots/aa/">`)+hl));
    }).catch(err => console.log(err));
 }
 Promise.all([loadSnapshot(s1URL, s1), loadSnapshot(s2URL, s2)]).then(() => {
     console.log("snapshots loaded");
 });
</script>
{{ end }}
