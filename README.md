Jishaku Web
===========

Jishaku Toshokan's web component!

HOW 2 GET STARTD
----------------

1. Download the package for your platform from [bintray](https://bintray.com/deoxxa/jishaku/web/_latestVersion)
2. Extract the .zip file
3. Run the `jishaku` program *from within that directory*
4. Visit [http://127.0.0.1:3000/](http://127.0.0.1:3000/)

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

Running
-------

This is pretty darned simple. Unzip the distribution package, then just run the
`jishaku` binary (or `jishaku.exe` if you're on Windows). Make sure you run it
from within that directory, or it won't be able to find the data files.

Hacking
-------

Changing static files requires no restarting. Changing a template requires a
restart of the server. Changing the server requires a complete rebuild.

Binaries
--------

Binaries are available from [bintray](https://bintray.com/deoxxa/jishaku/web/_latestVersion).
I'll try to keep them updated, but feel free to poke me if they get too stale.

Building
--------

Building is laborious and annoying, but the short version is that you have to
get bleve to build first, then `go build fknsrs.biz/p/jishaku-web/server`. To
build the distribution binaries, I've got some patches for bleve that remove a
lot of dependencies that aren't used for Jishaku. You can find those patches in
[my dev branch](https://github.com/deoxxa/bleve/tree/dev).

License
-------

3-clause BSD. A copy is included with the source.
