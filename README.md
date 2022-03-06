# Stop the war in Ukraine!
----

Personal goal reached: Know Go language and have an overview of what pure
HTML/CSS/JS can do! Web frontend definitley needs better structure.

Currently I'm trying to have a more modular structure of BC+ and
playing around with that. **This will reset new commits of BC+ to
unusable state** before I start with building new things up. 

Reworking the thing is currently in rather slow progress…

--------------------

![Logo](assets/s/img/Logo.png)

# BoardComputer+ for E:D

[Binary Downloads](https://github.com/CmdrVasquess/BCplus/releases) –
[Documentation Index](https://cmdrvasquess.github.io/bcplus/)

--------------------

BC+ evaluates the player's journal and serves web pages with useful
informaton. I.e. one can easily access BC+ from any computer, tablet
or smartphone without ATL-TAB'ing away from E:D.

Currnetly information for long distance tarvel is supported - I was on
the way to Beagle Point when I startet working on this. It displays
estimations based on the recent jump history along with your galactic
position:

![Travel Screen](docs/imgs/screen-travel.jpg?raw=true)

As well as the availability of raw materials on scanned bodies in the
current system:

![Materials Screen](docs/imgs/screen-mats.jpg?raw=true)

## Installation

BC+ currently is only provided for E:D on PC – though it compliles fine
on Linux (my dev platform). It can be run from any directory as long as
the directory with necessary `assets` is accessible in the same
directory.

If you download the binary distribution, just unpack the ZIP file. This
will create a BCplus folder containing all things you need.

## Running

With a standard E:D installation from Frontier it should be perfectly
fine to just double-click the BCplus.exe. It should find the journal
files in the standard location. If you don't have a standard filesystem
layout, there are some command-line options that let you change
directory paths and other things.

After running the program a web-server is running on your machine (that's
why the windows firewall will ask you if BCplus is permitted to access
the network the first time you run the program). You can open the web 
pages with your local browser on `http://localhost:1337`. If you want to
run the browser on another device in your network you can do so, if you
know how to address your E:D host, e.g. by IP like `192.168.0.2`. Then
you would enter `http://192.168.0.2:1337` in the browser on the other
device. 

### Options

First, option syntax is not Windows standard – BC+ is written in
[Go](https://golang.org) and uses Go's standard command line parsing
package. So be prepared to start options with '-' not '/'.

* `-j <directory>` set the path to the directory containing your journal
  files. (default: %HOME%\Saved Games\Frontier Developments\Elite Dangerous)

* `-p <port>` set the port on which the web server is listening (default:
  1337).

* `-d <directory>` set the directory where BC+ collects its data (BC+
  handles multiple commanders)

* `-h` show help information, i.e. the complete and up-to-date list of
  options.

## Building from Source

To build it BC+ source you need

1. A proper installation of the [Go SDK](https://golang.org). The download
   will take you directly to the installation instructions. On Windows the 
   installer should do the necessary things for you.

2. BC+ has/had/ will have dependencies that use [`cgo`](https://golang.org/cmd/cgo/).
   For `cgo` to work one needs a working C compiler. Details can be found on
   the `cgo` doc pages. I use [MinGW-w64](https://sourceforge.net/projects/mingw-w64/) 
   to build BC+. More details can be found on the
   [MinGW-w64 project page](http://mingw-w64.org/).

If you have things set up correctly ```go get github.com/CmdrVasquess/bcplus```
should build the executable without errors. However this does not create a
setup where `bcplus.exe` finds necessary resources. So, clone the repository,
`cd` into the directory and use `go build`. Now you can run directly from the
repo.

## Credits

1. Thanks to the [Go](https://golang.org) community to provide such a
   nice programming environment (BC+ primarily exists because I needed
   a project to learn Go).

2. Frontier Development for the nice game.

3. [Elite Dangerous Assets](http://edassets.org/) site for providing useful
   visual stuff.

4. The creators of [Vue.js](https://vuejs.org/) as it became an important part
   of the Web UI.

6. All the people who are giving us incredible technology stacks like the
   [Web](https://www.w3.org/), [OS'es](https://www.debian.org/) and
   [Tools](https://www.gnu.org/) and many many more without the ulterior
   motive of spying on us for their own profit. – Being it payed or free
   software, this addresses all who respect their users.

## Disclaimer

Board Computer plus was created using assets and imagery from
[Elite: Dangerous](https://www.elitedangerous.com/), with the
permission of [Frontier Developments plc](http://frontier.co.uk/), for
non-commercial purposes. It is not endorsed by nor reflects the views
or opinions of Frontier Developments and no employee of Frontier
Developments was involved in the making of it.
