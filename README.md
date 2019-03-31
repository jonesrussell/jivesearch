<p align="center">
  <a href="https://github.com/adamfaliq42/jivesearch/edit/master/README.md">
    <img alt="jive-search logo" src="frontend/static/icons/logo.png">
  </a>
</p>

<br>


<p align="center">
Jive Search is the open source search engine that doesn't track you. Search privately, now : https://jivesearch.com
</p>

<br>

<p align="center">
   <a href="https://github.com/jivesearch/jivesearch"><img src="https://img.shields.io/badge/go-1.12.1-blue.svg"></a>
   <a href="https://travis-ci.org/jivesearch/jivesearch"><img src="https://travis-ci.org/jivesearch/jivesearch.svg?branch=master"></a>
  <a href="https://github.com/jivesearch/jivesearch/blob/master/LICENSE"><img src="https://img.shields.io/badge/license-Apache-brightgreen.svg"></a>
</p>

<br>

## 💾 Installation
The below will build and run a Docker Compose file for Elasticsearch, OpenResty (Nginx), PostgreSQL, Redis, and a NSFW/image classification server. The OpenResty build assumes you have a domain name as well as a Let's Encrypt SSL certificate. However, in order for the nginx.conf file to dynamically load your SSL certificate you will need to create a symlink to a generic "domain" folder (replace "example.com" with your domain). For local development you can skip this step.

```bash
$ ln -s /etc/letsencrypt/live/example.com /etc/letsencrypt/live/domain
```

Install and run

```bash
$ go get -u github.com/jivesearch/jivesearch
$ cd $GOPATH/src/github.com/jivesearch/jivesearch/docker
$ domain=example.com && onion=myonion.onion && data_directory=/path/to/data && sudo mkdir -p $data_directory/elasticsearch && sudo chown 1000:1000 $data_directory/elasticsearch && sudo DATA_DIRECTORY=$data_directory REDIS_PORT=6379 ES_HEAP=2g NGINX_DOMAIN=$domain ONION=$onion docker-compose build --no-cache && sudo DATA_DIRECTORY=$data_directory REDIS_PORT=6379 ES_HEAP=2g NGINX_DOMAIN=$domain ONION=$onion docker-compose up -d
```

**For local development you don't need a Let's Encrypt certificate. You can simply access http://127.0.0.1:8000 directly by running ```cd ~/go/src/github.com/jivesearch/jivesearch/frontend && go run ./cmd/frontend.go --debug=true --provider=yandex --images_provider=pixabay```** 

Elasticsearch may give you an error about max virtual memory areas. In that case:
```bash
# Add the following to the bottom of /etc/sysctl.conf (requires restart):
vm.max_map_count=262144
```

For systemd settings (replace "myuser" below and edit env variables as needed):
```bash
sudo curl -o /etc/systemd/system/frontend.service 
https://gist.githubusercontent.com/brentadamson/7b8117347909cc38384fed589a3d785d/raw/94d25b0751767554b3f898ee492f3e03577d0e5f/frontend
```
```bash
sudo curl -o /etc/systemd/system/images.service https://gist.githubusercontent.com/brentadamson/daafa09f8d06eb401e0eb72c2b992261/raw/357e66de29d56739ae41d61cbe227d36819e0df4/images
```
```bash
sudo curl -o /etc/systemd/system/crawler.service https://gist.githubusercontent.com/brentadamson/0880ef548130f69c2537049a550be8e8/raw/42269dfcba6d86aba49bc56ffa7e60a9eb7ebdf3/crawler
```
(The crawler is only necessary if you don't use a third-party search provider like Yandex.)


##### Wikipedia Dump File
```bash
$ cd $GOPATH/src/github.com/jivesearch/jivesearch/instant/wikipedia/cmd/dumper && go run dumper.go --workers=2 --dir=/path/to/dump/files --wikipedia=true --wikidata=true --wikiquote=true --wiktionary=true --truncate=400 --delete=true
```
##### Location Data
Location data is not logged but is used for local weather

```bash
$ sudo add-apt-repository ppa:maxmind/ppa
$ sudo apt update && sudo apt install geoipupdate
$ sudo nano /usr/local/etc/GeoIP.conf
  AccountID 0
  LicenseKey 000000000000
  EditionIDs GeoLite2-City GeoLite2-Country
$ sudo crontab -e
  56 3 * * 4 /usr/bin/geoipupdate
```

##### Timezone
A timezone database is used to get the timezone from a lat/lon for our clock instant answer
```bash
$ go get -u github.com/evanoberholster/timezoneLookup
$ go install $GOPATH/src/github.com/evanoberholster/timezoneLookup/cmd/timezone.go
$ rm -Rf ~/timezone && mkdir ~/timezone && cd ~/timezone
$ cd ~/timezone && wget https://github.com/evansiroky/timezone-boundary-builder/releases/download/2018i/timezones.geojson.zip
$ cd ~/timezone && unzip ~/timezone/timezones.geojson.zip
$ cd ~/timezone/dist && timezone -json "combined.json" -db=timezone -type=memory
$ sudo rm -Rf /usr/share/timezone
$ sudo mkdir /usr/share/timezone
$ sudo mv timezone.snap.json /usr/share/timezone
```

##### MusicBrainz
Instructions for MusicBrainz (to show the album discography instant answer) can be found [here](https://gist.github.com/brentadamson/b711d5c9c4d974d6999876004f8bc1cd).

<br>


## 🚀 **Roadmap** 
### Our goal is to create a search engine that respects your privacy AND delivers great search results, instant answers, maps, image search, news, and more. 
  
Marked items indicate progress has been made in that category. There is much more to do in each area. Suggestions are welcome!
- [x] Privacy
- [x] !Bangs
- [x] Core Search Results & Distributed Crawler    
    - [ ] Advanced Search (exact phrase, dogs OR cats,  -cats, site/domain search, etc.)
    - [ ] Filetype search
    - [x] Language & Region
    - [ ] Phrase Suggester (a.k.a. "Did You Mean?")
    - [x] Proxy Links
    - [x] SafeSearch    
    - [ ] Time Search (past year/month/day/hour, etc.
    - [x] 3rd party search providers
        - [x] Yandex API
- [x] Autocomplete
- [x] Instant Answers
    - [x] Birthstone, camelcase, characters, coin toss, frequency, POTUS, prime, random, reverse, stats, user agent, etc. 
    - [x] Breach (a.k.a. have i been pwned)
    - [x] Discography/Music albums
    - [x] Economic stats (GDP, population)
    - [ ] Flight Info & Status
    - [x] JavaScript-based answers
        - [x] Basic calculator
            - [x] Mortgage, financial and other calculators
        - [x] CSS/JavaScript/JSON/etc minifier and prettifier
        - [x] Converters (foreign exchange, meters to feet, mb to gb, etc...)
    - [x] Maps
    - [x] Nutrition
    - [x] Package Tracking (UPS, FedEx, USPS, etc...)
    - [ ] Shopping
    - [x] Stack Overflow
    - [x] Stock Quotes & Charts    
    - [x] Weather
    - [x] WHOIS
    - [x] Wikipedia summary
    - [x] Wikidata answers (how tall is, birthday, etc.)
    - [x] Wikiquote
    - [x] Wiktionary    
    - [ ] Many more instant answers (including from 3rd party APIs)
    - [ ] Translate trigger words and answers to other languages
- [x] Image Search
- [ ] Video Search
- [ ] News
- [ ] Custom CSS Themes
- [x] Tor

<br>

## 📙 Documentation
Jive Search's documentation is hosted on GoDoc Page [here](https://godoc.org/github.com/jivesearch/jivesearch).

<br>

## 💬 Contributing
Want to contribute? Great! 

Search for existing and closed issues. If your problem or idea is not addressed yet, please open a new issue [here](https://github.com/jivesearch/jivesearch/issues/new).

You can also join us in our chatroom at https://discord.gg/cfxQkuh or contact us on Twitter @jivesearch.

<br>

## 📜 Copyright and License
Code and documentation copyright 2018 the [Jive Search Authors](https://github.com/jivesearch/jivesearch/graphs/contributors). Code and docs released under the [Apache License](https://github.com/jivesearch/jivesearch/blob/master/LICENSE).
