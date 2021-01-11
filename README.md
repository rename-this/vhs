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


<!-- TABLE OF CONTENTS -->
<details open="open">
  <summary><h2 style="display: inline-block">Table of Contents</h2></summary>
  <ol>
    <li><a href="#about-the-project">About The Project</a></li>
    <li><a href="#getting-started">Getting Started</a></li>
    <li><a href="#more-information">More Information</a></li>    
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#built-with">Built With</a></li>
    <li><a href="#contact">Contact</a></li>
  </ol>
</details>


## About The Project

`vhs` is a versatile tool for network traffic capture. `vhs` can be run as a command line tool or be deployed into your
Kubernetes cluster as a sidecar, where it can capture traffic to and from your services. Live traffic
can be analyzed in real time to produce Prometheus metrics or saved for use in offline analysis, load testing, or 
whatever you can imagine!


## Getting Started

The quickest way to see `vhs` in action is to visit our 
[getting started](https://rename-this.github.io/vhs/getting-started/) page to run through a simple demo of `vhs` and
see a few straightforward usage examples that demonstrate various aspects of `vhs` functionality.


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


## Roadmap

We want **your help** in determining the future of `vhs`. See the [issues](https://github.com/rename-this/vhs/issues)
page and please let us know what features are important to you!


## License

Distributed under the Apache 2.0 license. See `LICENSE` for more information.


## Built With
* [cobra](https://github.com/spf13/cobra)
* [zerolog](https://github.com/rs/zerolog)
* [go-packet](https://github.com/google/gopacket)


## Contact
There are several ways to contact the `vhs` team:
* You can file a GitHub [issue](https://github.com/rename-this/vhs/issues) with an idea, feature request, or bug.
* You can join the public `vhs` [mailing list](https://groups.google.com/g/vhs-pre-rename-launch).
* You can join the public `vhs` [Slack](https://stormforge.slack.com).

If you are interested in contributing, please join the mailing list and Slack above, and be sure to check out the 
[contributor guidelines](https://github.com/rename-this/vhs/blob/main/CONTRIBUTING.md) and the
[code of conduct](https://github.com/rename-this/vhs/blob/main/CODE_OF_CONDUCT.md).



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