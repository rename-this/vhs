<!-- PROJECT SHIELDS -->
[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![MIT License][license-shield]][license-url]



<!-- PROJECT LOGO -->
<br />
<p align="center">
<h1 align="center" >:vhs:</h1> 
<h3 align="center">VHS</h3>

  <p align="center">
    The cloud-native swiss army knife for network traffic data.
    <br />
    <a href="https://rename-this.github.io/vhs/"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://github.com/rename-this/vhs/issues">Report Bug</a>
    ·
    <a href="https://github.com/rename-this/vhs/issues">Request Feature</a>
  </p>
</p>



<!-- TABLE OF CONTENTS -->
<details open="open">
  <summary><h2 style="display: inline-block">Table of Contents</h2></summary>
  <ol>
    <li><a href="#about-the-project">About The Project</a></li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#demo-setup">Demo Setup</a></li>
        <li><a href="#run-the-demo">Run The Demo</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#built-with">Built With</a></li>
    <li><a href="#contact">Contact</a></li>
  </ol>
</details>



<!-- ABOUT THE PROJECT -->
## About The Project

`vhs` is a versatile tool for network traffic capture. `vhs` can be run as a command line tool or be deployed into your
Kubernetes cluster as a sidecar, where it can capture traffic to and from your services. Captured traffic
can be analyzed to produce live Prometheus metrics or saved for use in offline analysis, load testing, or whatever
you can imagine!

<!-- GETTING STARTED -->
## Getting Started

The quickest way to see `vhs` in action is to use the development Docker image in this repo to run a demo.

### Prerequisites
The first step is to clone the VHS repository.

```
git clone https://github.com/rename-this/vhs.git
```

You will need a working Docker installation to successfully execute the following commands. You should be able to
install Docker from your favorite package manager, or you can see the [Docker site](https://www.docker.com/get-started)
for more details.

Once you have Docker set up and working, you can build the development container (which installs some helpful tools
and mounts the source in the container) by changing directory into the repository you cloned and running:

```
$ make dev
```

### Demo Setup
Open a bash session on the container by running the following command in a terminal:

```
$ docker exect -it vhs_dev bash
```

In your new bash session, run this script to start a simple echo server and a curl request that calls the server every
second:

```
$ cd testdata && ./echo.bash
```

This will generate some local HTTP traffic for `vhs` to capture.

### Run The Demo
Open another bash session on the container in a new terminal:

```
$ docker exect -it vhs_dev bash
```

In this new session, run the following command to build `vhs` and run it to capture our local HTTP traffic.

```
$ go build ./cmd/vhs && ./vhs --input "tcp|http" --output "json|stdout" --capture-response --address 0.0.0.0:1111 --middleware ./testdata/http_middleware.bash | jq -R "fromjson | .body" 2> /dev/null
```

The output should look something like this:

```
"hello, world 1594678195 [[hijacked 0]]"
"hello, world 1594678195 [[hijacked 1]]"
"hello, world 1594678196 [[hijacked 0]]"
"hello, world 1594678196 [[hijacked 1]]"
"hello, world 1594678197 [[hijacked 0]]"
"hello, world 1594678197 [[hijacked 1]]"
"hello, world 1594678198 [[hijacked 0]]"
"hello, world 1594678198 [[hijacked 1]]"
"hello, world 1594678199 [[hijacked 0]]"
"hello, world 1594678199 [[hijacked 1]]"
"hello, world 1594678200 [[hijacked 0]]"
```
<!-- USAGE EXAMPLES -->
## Usage

### Capture live HTTP traffic to cloud storage
`./vhs --input "tcp|http" --output "json|gzip|gcs" --address 0.0.0.0:80 --capture-response --gcs-bucket-name <some_bucket> --gcs-object-name <some_object> --flow-duration 2m`

The above command will capture live HTTP traffic on port 80 for 2 minutes and save the captured data as gzipped JSON to
Google Cloud Storage (GCS) in the specified bucket and object.

### Generate HAR file from saved HTTP data
`./vhs --input "gcs|gzip|json --output "har|stdout" --gcs-bucket-name <some_bucket> --gcs-object-name <some_object> --flow-duration 15s > harfile.json`

The above command will "replay" saved HTTP data in gzipped JSON format from the specified GCS bucket and object and
produce an HTTP Archive ([HAR](http://www.softwareishard.com/blog/har-12-spec/)) file called `harfile.json` on the local
filesystem.

### Provide live HTTP metrics
`./vhs --input "tcp|http" --address 0.0.0.0:80 --capture-response --prometheus-address 0.0.0.0:8888 --flow-duration 60m`

The above command will capture live HTTP traffic on port 80 and calculate
[RED metrics](https://www.weave.works/blog/the-red-method-key-metrics-for-microservices-architecture/) for the captured
data in real time. These metrics will be available on a [Prometheus](https://prometheus.io/) endpoint at port 8888. The
command will run for 60 minutes.

### More Information
A [complete guide](https://rename-this.github.io/vhs/reference/) to the usage and flags of `vhs` is available on
our documentation site. `vhs` operates on the concept of a data flow that originates with a 
[source](https://rename-this.github.io/vhs/reference/#sources) and
terminates with one or more [sinks](https://rename-this.github.io/vhs/reference/#sinks). Sources may capture network
data, read data from files, etc. Sinks may write data to cloud (GCS, S3, etc.) or local storage, standard output, or
send data to other destinations. Along the way, data may pass through a series of input 
[modifiers](https://rename-this.github.io/vhs/reference/#input-modifiers) and 
[formats](https://rename-this.github.io/vhs/reference/#input-formats) and output 
[modifiers](https://rename-this.github.io/vhs/reference/#output-modifiers)
and [formats](https://rename-this.github.io/vhs/reference/#output-formats) that transform the data. This architecture
is described in more detail [here](https://rename-this.github.io/vhs/architecture/).

<!-- ROADMAP -->
## Roadmap

We want **your help** in determining the future of `vhs`. See the [issues](https://github.com/rename-this/vhs/issues)
page and please let us know what features are important to you!

<!-- LICENSE -->
## License

Distributed under the Apache 2.0 license. See `LICENSE` for more information.


## Built With
* [cobra](https://github.com/spf13/cobra)
* [zerolog](github.com/rs/zerolog)
* [go-packet](https://github.com/google/gopacket)


<!-- CONTACT -->
## Contact

Your Name - [@twitter_handle](https://twitter.com/twitter_handle) - email

Project Link: [https://github.com/rename-this/vhs](https://github.com/rename-this/vhs)


<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://img.shields.io/github/contributors/rename-this/vhs.svg?style=for-the-badge
[contributors-url]: https://github.com/rename-this/vhs/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/rename-this/vhs.svg?style=for-the-badge
[forks-url]: https://github.com/rename-this/vhs/network/members
[stars-shield]: https://img.shields.io/github/stars/rename-this/vhs.svg?style=for-the-badge
[stars-url]: https://github.com/rename-this/vhs/stargazers
[issues-shield]: https://img.shields.io/github/issues/rename-this/vhs.svg?style=for-the-badge
[issues-url]: https://github.com/rename-this/vhs/issues
[license-shield]: https://img.shields.io/github/license/rename-this/vhs.svg?style=for-the-badge
[license-url]: https://github.com/rename-this/vhs/blob/master/LICENSE.txt
[linkedin-shield]: https://img.shields.io/badge/-LinkedIn-black.svg?style=for-the-badge&logo=linkedin&colorB=555
[linkedin-url]: https://www.linkedin.com/company/stormforge/