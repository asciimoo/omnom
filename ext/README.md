# Omnom Extension

The Omnom extension project.
The project is pre-compiled with webpack. This gives the opportunity to separate the extension code into modules.
The main entry point for extension popup is the omnom.js, and for the options page it's options.js.


## Requirements

npm >= 6.14.15

## Setup & run

### Install Dependencies
Run ```npm i to install``` all dependencies. 
Install cross-env global npm package.

To install the extension in Chrome: 
- Open extension page (chrome://extensions/)
- Enable developer mode
- Click on Load Unpacked
- Find the build folder
- Select the folder

To install the extension in Firefox: 
- Open addons page (about:addons)
- Click on cog icon
- Go to Debug Add-ons (about:debugging#/runtime/this-firefox)
- Click on Load Temporary Add-on...
- Locate build folder
- Open manifest.json form build folder

After the extension installed, you can use it, and if you've run the project in watch mode you can see the changes in real time.
(Maybe you have to clode and reopen the extension to see the changes)

### Run & Build

#### Run
You can run the project in watch mode with ```npm run start```

In watch mode, every change in the extension src folder will trigger to rebuild the extension to the build folder.

#### Build
To build the project run ```npm run build```
This will build the extension with minified, uglified files to the ext/build folder.

## Bugs


## License

AGPLv3
