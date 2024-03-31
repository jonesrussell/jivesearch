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
   <a href="https://github.com/jonesrussell/jivesearch"><img src="https://img.shields.io/badge/go-1.12.1-blue.svg"></a>
   <a href="https://travis-ci.org/jonesrussell/jivesearch"><img src="https://travis-ci.org/jonesrussell/jivesearch.svg?branch=master"></a>
  <a href="https://github.com/jonesrussell/jivesearch/blob/master/LICENSE"><img src="https://img.shields.io/badge/license-Apache-brightgreen.svg"></a>
</p>

<br>

## 💾 Installation
```bash
go get -u github.com/jonesrussell/jivesearch
JIVESEARCH_YANDEX_USER="" && JIVESEARCH_YANDEX_KEY="" && JIVESEARCH_PIXABAY_KEY=""
cd ~/go/src/github.com/jonesrussell/jivesearch/frontend && go run ./cmd/frontend.go --debug=true --provider=yandex --images_provider=pixabay
```

A Yandex user/API key can be obtained here: https://tech.yandex.com/xml/

A Pixabay API key can be obtained here: https://pixabay.com/api/docs/

Other API keys and settings can likewise be set via environment variables: https://github.com/jonesrussell/jivesearch/blob/master/config/config.go

For production usage see https://github.com/jonesrussell/jivesearch/blob/bcf9c1e6e52cd2bc9fe7e97982509fe8288b41dc/README.md

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
Jive Search's documentation is hosted on GoDoc Page [here](https://godoc.org/github.com/jonesrussell/jivesearch).

<br>

## 💬 Contributing
Want to contribute? Great! 

Search for existing and closed issues. If your problem or idea is not addressed yet, please open a new issue [here](https://github.com/jonesrussell/jivesearch/issues/new).

You can also [join us on Slack](https://join.slack.com/t/jivesearch/shared_invite/enQtNTkwMjg1OTc3MjgyLWZiMDFjMWM5NGU4OWNmY2Q3YjUzZGMxN2ZiNDBmYWVhMzZkMzlmNThlNTE3ZjY1MTU5MDBhNDNkNDM0NmU2MmY) or contact us on Twitter @jivesearch.

<br>

## 📜 Copyright and License
Code and documentation copyright 2018 the [Jive Search Authors](https://github.com/jonesrussell/jivesearch/graphs/contributors). Code and docs released under the [Apache License](https://github.com/jonesrussell/jivesearch/blob/master/LICENSE).
