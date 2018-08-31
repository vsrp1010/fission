# Record Replay Debug

Description/Explanation of Record/Replay TBD.

### Debug

Current state: Development on the debug functionality is incomplete. 

These instructions are specific to the NodeJS/Visual Studio Code editor. 

# Steps To Run

1. Build a debug image to use

Sample Dockerfile:
```bash
FROM node:8-alpine

ARG NODE_ENV
ENV NODE_ENV $NODE_ENV

RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app

COPY package.json /usr/src/app/
RUN npm install && npm cache clean --force
COPY server.js /usr/src/app/server.js

CMD [ "node", "--inspect", "server.js" ]

EXPOSE 8888
```
Commands:
```sh
docker build -t <<username/debug-image>>
```
```sh
docker push <<username/debug-image>>
```


2. Create an environment referencing this debug image

```sh
fission env create --name nodejs --image fission/node-env:0.9.0 --debug-image <<username/debug-image>>
```

3. Create a function for this environment with a breakpoint set

Sample function, hello.js:
```javascript
debugger;
module.exports = async function(context) {
    return {
        status: 200,
        body: "Hello, world!\n"
    };
}
```
```sh
fission function create --name hello --code hello.js --env nodejs
```
4. Create a recorder for the function, issue a request that should get recorded

```sh
fission recorder create --name tracker --function hello
```
```sh
curl "http://$FISSION_ROUTER/fission-function/hello"
```
5. Identify the unique ID of the previously recorded request

```sh
fission records view --from 15s --to 0s
```

6. Get the internal cluster service IP running the debug image by requesting to replay and debug the above request

```sh
fission replay --reqUID <<reqUID>> --debug
```

# Planned TODO

* Get the command to return the POD NAME, not the service IP -- investigate Endpoint objects on the cluster on K8s docs
* Sometimes the debug image cannot be extracted from the environment; this happens sporadically but often enough to warrant investigation
* Set up a port-forwarding session for the user into the pod at the debugger-default port
* The user should edit a local launch.json file as follows. Ideally this should be auto-generated based on the environment spec of the relevant request.
* Set the initial breakpoint on behalf of the user at the user's function using the [debugger/inspector protocol with the JSON API](https://github.com/dtretyakov/node-tools/wiki/Debugging-Protocol#request-setbreakpoint)

```
{
    "version": "0.2.0",
    "configurations": [
        
        {
            "type": "node",
            "request": "attach",
            "name": "Launch Program",
            "localRoot": "${workspaceFolder}",
            "remoteRoot": "/usr/src/app",
            "address": "localhost",
            "port": 9229
        }
    ]
}
```


