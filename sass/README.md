# Omnom SASS

Sass project for omnom extension, and web app.
The project uses bulma, and some of it's extensions as a git submodule.

The output of the project is a compiled, minified css, which can be imported as a static resource.

## Requirements

npm >= 6.14.15

## Setup & run

### Install Dependencies
Run ```npm i to install``` all dependencies. 
Install copyfiles, rimraf as global npm package.
Run ```git submodule update --init --recursive```

### Run
Run ```npm run build:css:extension``` to build the project, and copy the completed css file to the static css folder and to the extension's css folder.

## Bugs


## License

AGPLv3
