## Realtime Persistent Replication Basic Example

#### Getting started (on UNIX)

##### Required ports

- Frontend - 9001
- Backend - 9999

##### Backend - Go, SQLite

* Install Go 1.21
* (optional) Install Docker
* Install dependencies

```shell
cd be && go mod download
```

* Run the app

```shell
cd be && ./run.sh
```

* or Run the app inside container

```shell
cd be && ./run_docker.sh
```

&nbsp;

##### Frontend - React, Vite, Typescript, NPM

* Download/install latest NodeJS/NPM
* Install dependencies

```shell
cd fe && npm install
```

* Run the app

```shell
cd fe && npm run dev
```

* [Open in browser](http://localhost:9001) (browser must support [Websocket](https://caniuse.com/websockets) and
  optionally [IndexedDB](https://caniuse.com/indexeddb) / [WebWorker](https://caniuse.com/webworkers))

&nbsp;

###### Limitations

- Backend is limited to 1 instance, opening more may result in database locking
- Frontend is limited to 1 tab, opening more would result in double fetching and may result in overwriting/bad-state
  IndexedDB

###### Schema structure

- table `kv` stores the data, `kv_changes` keeps a temporary "changelog" of INSERTs/UPDATEs/DELETEs which the app
  listens to and reacts (after changelog is processed it's deleted)
- table `kv_access` stores the access to `kv`, `kv_access_changes` keeps a temporary "changelog" of
  INSERTs/UPDATEs/DELETEs which the app listens to and reacts  (after changelog is processed it's deleted)

###### Structure

```
├── be
│   ├── db -- Database layer, schema bootstrap and reading/writing
│   │   ├── base.go
│   │   ├── change.go
│   │   ├── driver.go
│   │   ├── read.go
│   │   ├── sql
│   │   │   ├── kv_access.sql
│   │   │   ├── kv.sql
│   │   │   └── triggers.sql
│   │   ├── types.go
│   │   └── write.go
│   ├── docker-compose.yml
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── integration_change_test.go
│   ├── integration_test.go
│   ├── main.go
│   ├── main_test.go
│   ├── rpr -- RPR Layer, listens/sends and processes RPR requests
│   │   ├── base.go
│   │   ├── change.go
│   │   ├── transport.go
│   │   └── types.go
│   ├── run_docker.sh
│   ├── run.sh
│   ├── session
│   │   ├── mgmt.go
│   │   └── types.go
│   ├── transport -- Transport Layer, HTTP/WS
│   │   ├── add.go
│   │   ├── base.go
│   │   ├── edit.go
│   │   ├── remove_access.go
│   │   ├── remove.go
│   │   ├── util.go
│   │   └── websocket.go
│   └── util
│       ├── convert.go
│       └── convert_test.go
├── fe
│   ├── index.html
│   ├── package.json
│   ├── package-lock.json
│   ├── src
│   │   ├── App.css
│   │   ├── App.tsx
│   │   ├── components
│   │   │   ├── AddDialog.tsx
│   │   │   ├── EditDialog.tsx
│   │   │   └── RemoveDialogs.tsx
│   │   ├── index.css
│   │   ├── main.tsx
│   │   ├── rpr
│   │   │   ├── RPRConst.ts
│   │   │   ├── RPRCoordinator.ts - Recieves/sends request from/to WS and publishes create/update/remove events
│   │   │   ├── RPRIndexedDB.ts - Read/writes from/to IndexedDB
│   │   │   ├── RPRProtocol.ts - Type definitions
│   │   │   ├── RPR.ts - Manages the Webworker (where available)/fallbacks to standard
│   │   │   ├── RPRWorker.ts
│   │   │   └── RPRWS.ts - Transport, sending and receiving WS
│   │   ├── useLiveData.ts - Handy React hook
│   │   ├── utils
│   │   │   ├── SMap.ts
│   │   │   └── Utils.ts
│   │   └── vite-env.d.ts
│   ├── tsconfig.json
│   ├── tsconfig.node.json
│   └── vite.config.ts
└── README.md
```
