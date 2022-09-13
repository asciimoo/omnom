const puppeteer = require('puppeteer');
const assert = require('assert');
console.log(process.argv);
if(process.argv.length != 3) {
    console.log("[E] Invalud number of arguments. Server address required");
    process.exit(1);
}

let serverAddr = process.argv[2];
let extId = '';
let extBaseUrl = '';

function sleep(time) {
    return new Promise(function(resolve) {
        setTimeout(resolve, time)
    });
}

async function getExtensionID(browser) {
    const page = await browser.newPage();
    await page.goto('chrome://extensions', { waitUntil: 'load' });
    return await page.evaluate(async function() {
        return await new Promise(resolve => {
            resolve(document.querySelector('extensions-manager').shadowRoot.querySelector("#items-list").shadowRoot.querySelector("extensions-item").getAttribute("id"));
        });
    });
}

async function fillSettings(browser) {
    let page = await browser.newPage();
    await page.goto(extBaseUrl+'popup.html', {waitUntil: 'load'});
    await page.waitForSelector("#token");
    const tokenInput = await page.$("#token");
    await tokenInput.type("test");
    const serverInput = await page.$("#url");
    await serverInput.type(serverAddr);
    page.evaluate((btnSelector) => {
        document.querySelector(btnSelector).click();
    }, 'input[type="submit"]');
    //await page.$eval('form', form => form.submit());
    page = await browser.newPage();
    await page.goto(extBaseUrl+'popup.html', {waitUntil: 'load'});
    const titleInput = await page.$("#title");
    assert(titleInput && titleInput != null && titleInput != undefined);
    await page.screenshot({path: 'extension.png'});
}

const tests = [
    fillSettings,
];

async function runTests(page) {
    for (let testFn of tests) {
        try {
            await testFn(page);
        } catch(e) {
            console.error("TEST '"+testFn.name+"' FAIL: \n", e.stack);
            process.exit(1);
        }
    }
}

(async () => {
    // Path to extension folder
    const extPath = '../../../ext/build/';
    try {
        console.log('==>Open Browser');
        const browser = await puppeteer.launch({
            // Define the browser location
            // Disable headless mode
            headless: false,
            // Pass the options to install the extension
            args: [
                `--disable-extensions-except=${extPath}`,
                `--load-extension=${extPath}`,
                `--window-size=1024,1024`
            ],
        });
        //console.log("ID!!! ", await getExtensionID(browser));

        console.log('==>Navigate to Extension');
        extId = await getExtensionID(browser);
        console.log('==>Extension ID: ', extId);
        extBaseUrl = `chrome-extension://${extId}/`;
        // Take a screenshot of the extension page
        await runTests(browser);
        //console.log('==>Take Screenshot');
        //await page.screenshot({path: 'extension.png'});

        //console.log('==>Close Browser');
        await browser.close();
    }
    catch (err) {
        console.error(err);
    }
})();
