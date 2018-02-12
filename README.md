# Consulate

Consulate provides a normalized HTTP endpoint that responds with
200 / non-200 according to Consul checks specified in the requested URL.
This endpoint can then be composed with existing HTTP-URL-monitoring tools like
AWS ELB health checks to enable service QoS monitoring based on Consul checks.

### Building from source

To build Consulate from the source code yourself you need to have a working
Go environment with [version 1.9 or greater installed](http://golang.org/doc/install).

    $ mkdir -p $GOPATH/src/github.com/kadaan
    $ cd $GOPATH/src/github.com/kadaan
    $ git clone https://github.com/kadaan/consulate.git
    $ cd consulate
    $ ./build.sh

## License

Apache License 2.0, see [LICENSE](https://github.com/kadaan/consulate/blob/master/LICENSE).