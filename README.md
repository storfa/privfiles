# privfiles
Securely share files that will self-destruct after being downloaded.

## Getting Started
Build the client app:

```bash
$ pushd client
$ npm install
```

Install the bower dependencies:

```bash
$ bower install
$ popd
```

Build install server dependencies:

```bash
$ pushd server
$ go install
```

Build run the server app:

```bash
$ go run main/main.go
```


