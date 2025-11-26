## Classes

<dl>
<dt><a href="#Document">Document</a></dt>
<dd><p>Represents an HTML document with transformation capabilities</p>
</dd>
</dl>

## Functions

<dl>
<dt><a href="#getDomData">getDomData()</a> ⇒ <code>Object</code></dt>
<dd><p>Extracts complete DOM data from the current page</p>
</dd>
</dl>

<a name="Document"></a>

## Document
Represents an HTML document with transformation capabilities

**Kind**: global class  

* [Document](#Document)
    * [new Document(html, text, url, doctype, title, htmlAttributes)](#new_Document_new)
    * [.absoluteUrl(url)](#Document+absoluteUrl) ⇒ <code>string</code>
    * [.getDomAsText()](#Document+getDomAsText) ⇒ <code>string</code>
    * [.transformDom()](#Document+transformDom)
    * [.walkDOM(node)](#Document+walkDOM) ⇒ <code>Promise</code>
    * [.transformNode(node)](#Document+transformNode)
    * [.transformLink(node)](#Document+transformLink)
    * [.transformStyle(node)](#Document+transformStyle)
    * [.transformImg(node)](#Document+transformImg)
    * [.transformIframe(node)](#Document+transformIframe)
    * [.transformTemplate(node)](#Document+transformTemplate)
    * [.setUrl(node)](#Document+setUrl)
    * [.rewriteAttributes(node)](#Document+rewriteAttributes) ⇒ <code>Promise</code>

<a name="new_Document_new"></a>

### new Document(html, text, url, doctype, title, htmlAttributes)
Creates a Document instance


| Param | Type | Description |
| --- | --- | --- |
| html | <code>string</code> | The HTML content |
| text | <code>string</code> | The extracted text content |
| url | <code>string</code> | The document URL |
| doctype | <code>string</code> | The document type declaration |
| title | <code>string</code> | The document title |
| htmlAttributes | <code>Object</code> | Attributes from the HTML element |

<a name="Document+absoluteUrl"></a>

### document.absoluteUrl(url) ⇒ <code>string</code>
Converts a relative URL to absolute using document's base URL

**Kind**: instance method of [<code>Document</code>](#Document)  
**Returns**: <code>string</code> - Absolute URL  

| Param | Type | Description |
| --- | --- | --- |
| url | <code>string</code> | The URL to convert |

<a name="Document+getDomAsText"></a>

### document.getDomAsText() ⇒ <code>string</code>
Gets the complete document as text (doctype + HTML)

**Kind**: instance method of [<code>Document</code>](#Document)  
**Returns**: <code>string</code> - The complete document text  
<a name="Document+transformDom"></a>

### document.transformDom()
Transforms the DOM by processing all nodes and downloading resources

**Kind**: instance method of [<code>Document</code>](#Document)  
<a name="Document+walkDOM"></a>

### document.walkDOM(node) ⇒ <code>Promise</code>
Recursively walks and transforms the DOM tree

**Kind**: instance method of [<code>Document</code>](#Document)  
**Returns**: <code>Promise</code> - Promise that resolves when all nodes are processed  

| Param | Type | Description |
| --- | --- | --- |
| node | <code>Node</code> | The node to process |

<a name="Document+transformNode"></a>

### document.transformNode(node)
Transforms a single node based on its type

**Kind**: instance method of [<code>Document</code>](#Document)  

| Param | Type | Description |
| --- | --- | --- |
| node | <code>Node</code> | The node to transform |

<a name="Document+transformLink"></a>

### document.transformLink(node)
Transforms LINK elements (stylesheets, icons, preloads)

**Kind**: instance method of [<code>Document</code>](#Document)  

| Param | Type | Description |
| --- | --- | --- |
| node | <code>HTMLLinkElement</code> | The link element to transform |

<a name="Document+transformStyle"></a>

### document.transformStyle(node)
Transforms STYLE elements by sanitizing CSS

**Kind**: instance method of [<code>Document</code>](#Document)  

| Param | Type | Description |
| --- | --- | --- |
| node | <code>HTMLStyleElement</code> | The style element to transform |

<a name="Document+transformImg"></a>

### document.transformImg(node)
Transforms IMG elements by downloading and embedding images

**Kind**: instance method of [<code>Document</code>](#Document)  

| Param | Type | Description |
| --- | --- | --- |
| node | <code>HTMLImageElement</code> | The image element to transform |

<a name="Document+transformIframe"></a>

### document.transformIframe(node)
Transforms IFRAME elements by embedding their content

**Kind**: instance method of [<code>Document</code>](#Document)  

| Param | Type | Description |
| --- | --- | --- |
| node | <code>HTMLIFrameElement</code> | The iframe element to transform |

<a name="Document+transformTemplate"></a>

### document.transformTemplate(node)
Transforms TEMPLATE elements by processing their content

**Kind**: instance method of [<code>Document</code>](#Document)  

| Param | Type | Description |
| --- | --- | --- |
| node | <code>HTMLTemplateElement</code> | The template element to transform |

<a name="Document+setUrl"></a>

### document.setUrl(node)
Processes BASE elements to update the base URL

**Kind**: instance method of [<code>Document</code>](#Document)  

| Param | Type | Description |
| --- | --- | --- |
| node | <code>HTMLBaseElement</code> | The base element to process |

<a name="Document+rewriteAttributes"></a>

### document.rewriteAttributes(node) ⇒ <code>Promise</code>
Rewrites node attributes (sanitize event handlers, resolve URLs, sanitize styles)

**Kind**: instance method of [<code>Document</code>](#Document)  
**Returns**: <code>Promise</code> - Promise that resolves when all attributes are processed  

| Param | Type | Description |
| --- | --- | --- |
| node | <code>HTMLElement</code> | The element whose attributes to rewrite |

<a name="getDomData"></a>

## getDomData() ⇒ <code>Object</code>
Extracts complete DOM data from the current page

**Kind**: global function  
**Returns**: <code>Object</code> - Object containing HTML, text, title, URL, and other metadata  
**Properties**

| Name | Type | Description |
| --- | --- | --- |
| html | <code>string</code> | The serialized HTML content |
| attributes | <code>Object</code> | HTML element attributes |
| title | <code>string</code> | Page title |
| doctype | <code>string</code> | Document type declaration |
| iframeCount | <code>number</code> | Number of iframes in the page |
| url | <code>string</code> | Page URL |
| text | <code>string</code> | Extracted visible text content |

