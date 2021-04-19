var br = chrome;
var omnom_url = '';
var omnom_token = '';
var site_url = '';

function saveBookmark(e) {
	e.preventDefault();
    createSnapshot().then(async (result) => {
        //var win = window.open("", "omnomDebug", "menubar=yes,location=yes,resizable=yes,scrollbars=yes,status=yes");
        //if(win) {
        //    win.document.write(result);
        //}
        //console.log(result);
        //return
        var form = new FormData(document.forms['add']);
        form.append("snapshot", result);
        fetch(omnom_url+'add_bookmark', {
            method: 'POST',
            body: form,
            //headers: {
            //    'Content-type': 'application/json; charset=UTF-8'
            //}
        })
        .then(resp => {
            if(resp.status !== 200) {
                throw Error(resp.statusText);
            }
            window.close()
        })
        .catch(err => {
            document.body.innerHTML = '<h1>Failed to save bookmark: '+err+'</h1>';
        });
    });
}

function displayPopup() {
    document.querySelector("form").addEventListener("submit", saveBookmark);
    br.storage.local.get(['omnom_url', 'omnom_token'], function(data) {
        omnom_url = data.omnom_url || '';
        omnom_token = data.omnom_token || '';
        if(omnom_url == '') {
            document.body.innerHTML = '<h1>Server URL not found. Specify it in the extension\'s options</h1><p>(right click on the extension\'s icon and select "options")';
            return;
        }
        if(omnom_token == '') {
            document.body.innerHTML = '<h1>Token not found. Specify it in the extension\'s options</h1><p>(right click on the extension\'s icon and select "options")';
            return;
        }
        document.getElementById("omnom_url").innerHTML = "Server URL: "+omnom_url;
        document.querySelector("form").action = omnom_url+'add_bookmark';
        document.getElementById("token").value = omnom_token;
        // fill url input field
        var url;
        br.tabs.query({active: true, lastFocusedWindow: true}, (tabs) => {
            document.getElementById('url').value = tabs[0].url;
            site_url = tabs[0].url;
        });
        // fill title input field
        br.tabs.executeScript({
            code: 'document.title;'
        }, (title) => {
            if(title && title[0]) {
                document.getElementById('title').value = title[0];
            }
        });
        // fill notes input field
        br.tabs.executeScript( {
            code: "window.getSelection().toString();"
        }, function(selection) {
            if(selection && selection[0]) {
                document.getElementById("notes").value = selection[0];
            }
        });
    });
}

function rewriteAttributes(node) {
    for(var i=0; i <  node.attributes.length; i++) {
        var attr = node.attributes[i];
        if(attr.nodeName === undefined) {
            continue;
        }
        if(attr.nodeName.startsWith("on")) {
            attr.nodeValue = '';
        } else if(attr.nodeName.startsWith("data-")) {
            attr.nodeValue = '';
        } else if(attr.nodeValue.startsWith("javascript:")) {
            attr.nodeValue = '';
        }
        if(attr.nodeName == "href") {
            attr.nodeValue = fullURL(attr.nodeValue);
        }
    }
}

function getDOMData() {
    function walkDOMS(node1, node2, func) {
        func(node1, node2);
        var children1 = node1.childNodes;
        var children2 = node2.childNodes;
        for (var i = 0; i < children1.length; i++) {
            if(children1[i].nodeType != Node.ELEMENT_NODE) {
                continue;
            }
            walkDOMS(children1[i], children2[i], func);
        }
    }
    function fullURL(url) {
        return new URL(url, window.location.origin).href
    }
    function getCSSText(obj) {
        if(obj.cssText) {
            return obj.cssText;
        }
        var text = '';
        for(var i=0; i < obj.length; i++) {
            var key = obj.item(i);
            text += key + ':' + obj[key] + ';';
        }
        return text;
    }
    var body = document.getElementsByTagName('body')[0];
    var ret = {
        'body': body.cloneNode(true),
        'title': '',
        'doctype': '',
        'htmlStyle': '',
    };
    if(document.doctype) {
        ret.doctype = new XMLSerializer().serializeToString(document.doctype);
    }
    if(document.getElementsByTagName('title').length > 0) {
        ret.title = document.getElementsByTagName('title')[0].innerText;
    }
    if(document.getElementsByTagName('html').length > 0) {
        ret.htmlStyle = getCSSText(window.getComputedStyle(document.getElementsByTagName('html')[0]));
    }
    var nodesToRemove = [];
    var hiddenStyleKeys = [
        'width',
        'height',
        'margin',
        'padding',
        'minWidth',
        'maxWidth',
        'minHeight',
        'maxHeight',
    ];
    walkDOMS(body, ret['body'], async function(n1, n2) {
        if(n1.nodeName == 'SCRIPT' || n1.nodeName == 'STYLE') {
            nodesToRemove.push(n2);
            return;
        }
        // TODO optimize styles - build css from inline styles
        n2.style = getCSSText(window.getComputedStyle(n1));
        // compute dynamic height/width below
        // TODO it's far from good..
        //var display = n1.style.display;
        //n1.style.display = "none";
        //var s = window.getComputedStyle(n1);
        //for(i in hiddenStyleKeys) {
        //    var k = hiddenStyleKeys[i];
        //    n2.style[k] = s[k];
        //}
        //n1.style.display = display;
    });
    for(i in nodesToRemove) {
        nodesToRemove[i].remove();
    }
    ret.body = ret.body.outerHTML;
    return ret;
}

async function createSnapshot() {
    var doc;
    function getDOM() {
        return new Promise((resolve, error) => {
            br.tabs.executeScript({
                code: '('+getDOMData+')()'
            }, (data) => {
                if(data && data[0]) {
                    doc = data[0];
                    resolve('');
                } else {
                    error('meh');
                }
            });
        });
    }
    await getDOM();
    var dom = document.createElement('html');
    var body = document.createElement('html');
    body.innerHTML = doc.body;
    await walkDOM(body, async function(node) {
        if(node.nodeType === Node.ELEMENT_NODE) {
            rewriteAttributes(node);
            if(node.nodeName == 'IMG') {
                node.src = await inlineFile(node.getAttribute("src"));
            }
        }
    });
    dom.appendChild(body)
    dom.style = doc.htmlStyle;
    dom.style.width = '100%';
    return doc.doctype+dom.outerHTML;
}

async function walkDOM(node, func) {
    await func(node);
    var children = node.childNodes;
    for (var i = 0; i < children.length; i++) {
        await walkDOM(children[i], func);
    }
}

function rewriteAttributes(node) {
    for(var i=0; i <  node.attributes.length; i++) {
        var attr = node.attributes[i];
        if(attr.nodeName === undefined) {
            continue;
        }
        if(attr.nodeName.startsWith("on")) {
            attr.nodeValue = '';
        }
        if(attr.nodeValue.startsWith("javascript:")) {
            attr.nodeValue = '';
        }
        if(attr.nodeName == "href") {
            attr.nodeValue = fullURL(attr.nodeValue);
        }
    }
}

async function inlineFile(url) {
    if(!url || (url || '').startsWith('data:')) {
        return url;
    }
    url = fullURL(url);
    console.log("fetching "+url);
	var options = {
	  method: 'GET',
	  cache: 'default'
	};
	var request = new Request(url);

	var imgObj = await fetch(request, options).then(async function(response) {
        var contentType = response.headers.get("Content-Type");
	    var buff = await response.arrayBuffer()
        var base64Flag = 'data:'+contentType+';base64,';
        return base64Flag + arrayBufferToBase64(buff);
	});
    return imgObj;
}

function arrayBufferToBase64(buffer) {
  var binary = '';
  var bytes = [].slice.call(new Uint8Array(buffer));
  bytes.forEach((b) => binary += String.fromCharCode(b));

  return btoa(binary);
}

function fullURL(url) {
    return new URL(url, site_url).href
}

document.addEventListener('DOMContentLoaded', displayPopup);

/* ---------------------------------*
 * End of omnom code                *
 * ---------------------------------*/
