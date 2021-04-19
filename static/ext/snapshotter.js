function walkTheDOM(node, func) {
    func(node);
    node = node.firstChild;
    while(node) {
        walkTheDOM(node, func);
        node = node.nextSibling;
    }
}

function createSnapshot() {
    br.tabs.executeScript({
        code: "document.getElementsByTagName('html')[0].cloneNode(true);"
    }, (dom) => {
        if(dom && dom[0]) {
            dom = dom[0];
            walkTheDOM(dom, function(node) {
                if (node.nodeType === 3) {
                    var text = node.data.trim();
                    if (text.length > 0) {
                        console.log(text);
                    }
                }
            });
        }
    });

}

