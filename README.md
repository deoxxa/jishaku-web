Jishaku Web
===========

Jishaku Toshokan's web component!

Overview
--------

This is the code that runs [jishaku.net](http://www.jishaku.net/). It's written
in [go](http://golang.org/) and should run just about anywhere. It's designed to
be usable both as a public system, where many users are being served by one
instance, and as a private system, where only a small group or a single user is
using it.

Configuration
-------------

There are two backends that can be used for storage, [ElasticSearch](http://www.elasticsearch.org/)
and [bleve](www.blevesearch.com). ElasticSearch requires an external server, but
may be more performant on large datasets. bleve is the default option, and does
not require an external server.

Binaries
--------

Prebuilt binaries will be available soon, but for now you'll have to build the
application. This will require some knowledge of the go build process.

Building
--------

Instructions pending.

License
-------

3-clause BSD. A copy is included with the source.
