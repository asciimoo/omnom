{{ define "full-content" }}
<div class="iframe-diff-wrapper">
    <h2 class="title">Side by side snapshot diff of {{ Truncate .SURL 250 }}</h2>
    <div class="container is-fluid content"><a href="{{ URLFor "Snapshot diff side by side" }}?s1={{ .S2.Key }}&s2={{ .S1.Key }}">swap sides</a></div>
    <noscript>{{ block "warning" KVData "Warning" "this feature requires javascript" "Tr" .Tr }}{{ end }}</noscript>
    <div class="columns">
        <div class="column">
            <a href="{{ URLFor "Snapshot" }}?sid={{ .S1.Key }}&bid={{ .S1.BookmarkID }}" class="is-size-3">{{ .S1.CreatedAt | ToDate }}</a>
            <iframe src="{{ SnapshotURL .S1.Key }}" id="sn1" title="snapshot 1" scrolling="no" class="snapshot-iframe"></iframe>
        </div>
        <div class="column">
            <a href="{{ URLFor "Snapshot" }}?sid={{ .S2.Key }}&bid={{ .S2.BookmarkID }}" class="is-size-3">{{ .S2.CreatedAt | ToDate }}</a>
            <iframe title="snapshot 2" id="sn2" class="snapshot-iframe" scrolling="no"></iframe>
        </div>
    </div>
</div>
<script id="highlighter" nomodule>
 function findNthString(s, n) {
     return containsNthString(s, document.body, n);
 }

 function occurrences(string, subString, allowOverlapping) {
     string += "";
     subString += "";
     if (subString.length <= 0) return (string.length + 1);

     var n = 0,
         pos = 0,
         step = allowOverlapping ? 1 : subString.length;

     while (true) {
         pos = string.indexOf(subString, pos);
         if (pos >= 0) {
            ++n;
             pos += step;
         } else break;
     }
     return n;
 }

 function nthIndex(str, pat, n){
     var L= str.length, i= -1;
     while(n-- && i++<L){
         i= str.indexOf(pat, i);
         if (i < 0) break;
     }
     return i;
 }

 function containsNthString(s, el, n) {
     if(occurrences(el.textContent, s) < n) {
         return null;
     }
     let count = 0;
     for(let c of el.children) {
         let o = occurrences(c.textContent, s)
         if(count + o >= n) {
             return containsNthString(s, c, n-count);
         }
         count += o;
     }
     let idx = nthIndex(el.textContent, s, n);
     return [el, idx];
 }

 function getStartPos(root, startPos) {
     for(c of root.childNodes) {
         if(c.nodeType == Node.TEXT_NODE) {
             let len = c.nodeValue.length;
             if(startPos < len) {
                 return [c, startPos];
             }
             startPos -= len;
         } else if(c.nodeType == Node.ELEMENT_NODE) {
             let sp = getStartPos(c, startPos);
             if(Array.isArray(sp)) {
                 return sp;
             }
             startPos = sp;
         }
     }
     return startPos;
 }

 function highlightString(s, n) {
     let f = findNthString(s, n)
     if(f == null) {
         return;
     }
     edits = [];
     while(s) {
         let sp = getStartPos(f[0], f[1]);
         if(sp == null) {
             break;
         }
         let hl = document.createElement("mark");
         hl.style = `background: #55efc4 !important; color: black !important;`;
         let st = sp[0].nodeValue;
         hl.innerText = st.slice(sp[1], sp[1]+s.length);
         edits.push([sp[0], [st.slice(0, sp[1]), hl, st.slice(sp[1]+s.length, sp[1].length)]]);
         s = s.slice(hl.textContent.length, s.length);
         f[1] += hl.textContent.length;
     }
     for(e of edits) {
         e[0].replaceWith(...e[1]);
     }
 }

 function insertString(beforeStr, n, s) {
     let f = findNthString(beforeStr, n)
     let sp = getStartPos(f[0], f[1]+beforeStr.length);
     let st = sp[0].nodeValue;
     let hl = document.createElement("mark");
     hl.style = `background: #ff9999 !important; color: black !important;`;
     hl.innerText = s;
     sp[0].replaceWith(st.slice(0, sp[1]), hl, st.slice(sp[1], sp[1].length));
 }

 let additions = {{ .Additions }};
 let deletions = {{ .Deletions }};
 for(d of deletions.reverse()) {
     insertString(d["preStr"], d["idx"], d["s"]);
 }
 for(a of additions) {
     highlightString(a["s"], a["idx"]);
 }
 //insertString("yolo", 1, "bolo");
 //highlightString("hai world", 3);
</script>
<script>
 var s1 = document.getElementById('sn1');
 var s2 = document.getElementById('sn2');
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
 s1.contentWindow.addEventListener('DOMContentLoaded', function(e) {
     resizeIFrameToFitContent(s1);
 });
 s2.addEventListener('load', function(e) {

     resizeIFrameToFitContent(s2);
 });

 let s2URL = "{{ SnapshotURL .S2.Key }}";
 fetch(s2URL).then(r => {
     return r.text();
 }).then(html => {
    let hl = document.getElementById("highlighter").cloneNode(true);
     hl.removeAttribute("nomodule");
     s2.srcdoc = html.replace("<head>", `<head><base href="./static/data/snapshots/aa/">`)+hl.outerHTML;
     //d.open();
     //d.write();
     //d.close();
     //s1.setAttribute("src", "data:text/html;base64,"+btoa(html.replace("<head>", `<head><base href="./static/data/snapshots/aa/">`)+hl));
 }).catch(err => console.log(err));
</script>
{{ end }}
