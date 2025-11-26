# Omnom-ext JS package

Omnom-ext is the standalone JS module of the omnom extension.

## Install

`npm i @asciimoo/omnom-ext`

## Usage

```JavaScript
import { getDomData } from "omnom-ext";

const data = getDomData();
console.log(data.html, data.text, data.url, data.doctype, data.title, data.attributes);
```

see [API_DOCS.md](API documentation)

## License

AGPLv3
