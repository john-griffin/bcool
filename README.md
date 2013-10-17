bcool
=====

Enhance the [bleedingcool.com][1] RSS feed with images and video.

## Synopsis

The [Bleeding Cool][1] news website only provides an abridged RSS feed. This web
app, written in Go, takes the original feed and enriches it with all the content
of the source page. This includes images, video and more text.

## Features

* Uses goroutines and channels to fetch each page in parallel. Very fast.
* No external dependancies.

## Usage

To compile and run locally, ensure you have [Go](http://golang.org), clone
the repo and run this command

``` bash
$ PORT=5000 go run bcool.go
```

Then navigate to `http://localhost:5000/feed` in your browser or aggregator to
see the feed.

To run your own hosted version it's really easy to deploy to Heroku by running

```bash
$ heroku create -b https://github.com/kr/heroku-buildpack-go.git
$ git push heroku master
```

## To do

The approach I have taken here could be applied generally to other feeds. I'd
like to extract this out and use it elsewhere.

[1]: http://www.bleedingcool.com
