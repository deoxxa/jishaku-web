# Help

---

### Topics

* [Frequently Asked Questions](#frequently-asked-questions)
  * [What is Jishaku?](#what-is-jishaku-)
  * [Hold on, hasn't Jishaku been around for a while?](#hold-on-hasn-t-jishaku-been-around-for-a-while-)
  * [Who are you?](#who-are-you-)
* [Logging In](#logging-in)
  * [Why should I log in?](#why-should-i-log-in-)
  * [Why don't I have a Jishaku account?](#why-don-t-i-have-a-jishaku-account-)
* [Searching](#searching)
  * [Search accuracy](#search-accuracy)
  * [Advanced search syntax](#advanced-search-syntax)

# Frequently Asked Questions

---

### What is Jishaku?

Jishaku is, first and foremost, an archive of bittorrent information. It exposes
a very flexible search mechanism for querying its database of bittorrent
metadata. No actual content is hosted on Jishaku, nor are the .torrent files
themselves. Jishaku is designed to serve as a historical archive and aggregation
of links to different sources for bittorrent files.

### Hold on, hasn't Jishaku been around for a while?

Correct. Jishaku was born out of frustration in late 2009, when the
then-dominant bittorrent index [Tokyo Toshokan](https://www.tokyotosho.info/)
went offline for an extended period. At that point, some friends and I (from the
now-defunct nekosubs fansubbing group) hastily put together a replacement, and
called it Jishaku Toshokan (as we were focussing on magnet links, and Jishaku
means magnet in Japanese.)

Over the years, Jishaku stagnated. Burdened with features and technical debt, it
began to collapse under its own weight until the day where it simply stopped
returning responses.

There was a constant background desire for me to restore it to its former glory,
until I finally found myself with a couple of spare days to put together a
minimally-functional prototype of what I wanted the new Jishaku to be - simple,
clean, and fast. By stripping away most of what made the previous Jishaku so
difficult to maintain and computationally demanding, an overwhemingly simple
core was exposed, and it's that core that will carry Jishaku into the future.

### Who are you?

Hi there! My name is Conrad, but you can call me CJ if you like. A lot of my
friends call me that. I live and work in Melbourne, Australia. I like
travelling, studying languages (both human and programming,) meeting people and
talking. Sometimes I get to do all these things at once!

On the internet, I'm most commonly known by the handle deoxxa:

* [twitter](https://twitter.com/deoxxa)
* [github](https://github.com/deoxxa)
* [email](mailto:deoxxa@fknsrs.biz)

# Logging In

---

### Why should I log in?

Logging in will record you as the owner of any submissions you make.

There are other features planned that will only work properly after logging in:

* Editing your submissions' descriptions and other metadata
* Removing submissions
* Personalised default searches
* Optional, per-user index pages

### Why don't I have a Jishaku account?

Keeping passwords and user information is really hard to get right. I don't want
the responsibility of keeping your personal information safe, so I've decided to
let Twitter take care of it for now. They seem to have a handle on it.

# Searching

---

### Search accuracy

The search engine powering Jishaku is designed to provide highly accurate search
results. If you have trouble finding a specific result, please [let me know](mailto:deoxxa@fknsrs.biz?subject=Jishaku Search Accuracy).

### Simple search syntax

By default, searching on Jishaku will just search in the title/comment of the
torrents. To search other fields, use the [advanced search syntax](#advanced-search-syntax).

### Advanced search syntax

Jishaku supports [ElasticSearch Query Syntax](http://www.elasticsearch.org/guide/en/elasticsearch/reference/current/query-dsl-query-string-query.html#query-string-syntax).
This advanced syntax allows you to construct very rich searches to hone in on
exactly what you want to find.

This functionality is triggered by adding **"\` "** to the start of your search
query (that's a grave accent, usually located next to the 1 key on your
keyboard, followed by a space.)
