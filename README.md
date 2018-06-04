# Getting Started

## Installation

Either:

* Clone this repository to `$GOPATH/src/github.com/cjcormack/file-watcher`, or,
* run `go get github.com/cjcormack/file-watcher/cmd/master`.

## Building

* Run `go install ...`.

## Running & Configuration

This solution has two daemons, `master` and `watcher`.

### `master`

This daemon provides two HTTP interfaces, one that is used for the public interface and the other that is
used for the `watcher`s to make a websocket connection to. It has the following command line args:

* `--watcherAddr`—this provides the address that the daemon listens on for connections from the watchers. Default `"localhost:8080"`.
* `--publicAddr`—this provides the address that the daemon listens on for the public API. Default `"localhost:8090"`.

### `watcher`

This daemon opens a websocket connection to the `master` and sends notifications about changes to the files listed
in the specified "watched" folder. It has the following command line args:

* `--masterAddr`—this provides the address to use when connecting to `master`. Default `"localhost:8080"`.
* `--folder`—this provides the path to the folder to watch.

# Discussion

## Use of websockets

I decided to use websockets as it seemed to give me a number of advantages for free:

* state—the `master` would easily be able to identify which `watcher` had sent an updated file list`,
* lifecycle—the `master` would be able to tell when a particular `watcher` had gone offline,
* synchronous—the `master` would be able to trust the order of the details that a given `watcher` transmitted,
* robustness—the `watcher`s would be able to easily tell when the `master` went offline and be able to transmit the
  current file list when the `master` becomes available again. This would remain true even if the master got restarted
  in the gaps between poll websocket messages.
* two-way—if we wanted the `master` to be able to communicate back to the `watcher`s this would avoid all of the complexities
  involved if a simple HTTP-based API was used

Most of my effort was focused on getting this aspect of the solution working well.

## Possible Improvements

The following are things that I would have given more attention to if more time was available:

* Tests. There are currently no tests. This is deeply upsetting.
* Configuration. The current use of command line args is not really suitable for a production system. If given more time
  I would have liked to have spent time adding support for an `.ini` based configuration.
* Scalability. I have not had a chance to investigate how well the architecture handles a lot of running watchers. There
  is partial support for each `watcher` watching multiple folders, and I anticipated that this, when mixed with multiple
  `watcher`s, would have given the platform a lot of capacity. (If this was a proper submission, I'd have probably aimed
  to finish the multiple folders bit).
* Efficiency. At the moment, the `watcher`s always send the full file list. To aid handling large folders with small change
  counts, it would be useful to send a diff containing the changes.
* Authentication. Currently the `master` and `watcher`s blindly trust each other.
* Plain text. Currently all of the data is transmitted in plain text (well JSON) over HTTP.
* Discovery. Getting the `watcher`s to automatically discover the `master` would have been an interesting bit to investigate.
* gRPC. It would have been intersting experimenting with gRPC as a replacement for the websocket aspect of the solution.
I was planning on refactoring things to make it easy to switch in a different transport.
* Use of consts. Currently in a few places I have left some uses of magic values, which I'd have liked to have dealt with (if
this was a proper submission, I would have sorted this out!).
* Refactoring. There're a couple of places where I'd have been a lot happier if I'd given the code a further further refactor
(again, I'd have sorted this out if this was a real submission!).
* Robustness. While I am hopeful that I have made some aspects of this solution robust enough, there are a few things
  that I haven't tested and I suspect at least one of them will break things (once again, I would have looked to fix these
  if it was a real submission):
** Watched folder deleted while `watcher` running—I am quite sure that something will break badly here,
** Folder name with spaces—I am pretty hopeful that this will work fine,
** Folder supplied as absolute path—I am hopeful, but wouldn't be surprised either way,
** Stale connections on `master`—I am not quite sure how this would be triggerable (if my testing the `master` was pretty
   good at noticing that the `watcher` had disconnected). Basically, I'd like to add some form of polling from `master` to
   the `watcher`s, possibly as a "I haven't had an update it a while. Could you send me the latest?" type request (which
   would kill the connection if it doesn't get a timely response).
