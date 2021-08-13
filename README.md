# Twitter-GO server

## Dependencies
```shell
$ go get -v github.com/dghubble/go-twitter/twitter 
$ go get -v github.com/dghubble/oauth1 
$ go get -v github.com/joho/godotenv
$ go get -v github.com/pusher/pusher-http-go
```
## Build

```shell
$ go run main.go
```

## Options

You can also make use of the trending topics on your Twitter if you want to. To search Twitter for other polls, you can also make use of the following command:

```shell
$ go run main.go -options="Apple,Javascript,Trump"
```

## Web

You will need to visit http://localhost:1500 to see the chart.

