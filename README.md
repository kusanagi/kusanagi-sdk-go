Go SDK for the KUSANAGI framework
=================================

[![Go Report Card](https://goreportcard.com/badge/github.com/kusanagi/kusanagi-sdk-go)](https://goreportcard.com/report/github.com/kusanagi/kusanagi-sdk-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

**Go** SDK to interface with the **KUSANAGI**™ framework (http://kusanagi.io).

Requirements
------------

* [KUSANAGI framework](http://kusanagi.io) 3.0+
* [Go](https://golang.org/dl/) 1.15+
* [libzmq](http://zeromq.org/intro:get-the-software) 4.2.5+

Installation
------------

Install the SDK using the following command:

```
$ go get github.com/kusanagi/kusanagi-sdk-go/v3@epoch-3
```

Getting Started
---------------

See the [getting started](http://kusanagi.io/docs/getting-started) tutorial to begin with the **KUSANAGI**™ framework and the **Go** SDK.

### Cancellation Signal

The Go SDK implements support to signal deadlines or cancellation through a read only channel that is available to **Middleware** and **Service** components.

The channel can be read using the `Api.Done()` method. For example, within a service action:

```go
func handler(action *kusanagi.Action) (*kusanagi.Action, error) {
    // Create a context for the current service call
    ctx, cancel := context.WithCancel(context.Background())

    // Create a channel to get the task result
    result := make(chan int)

    // Run some async task
    go task(ctx, result)

    // Cancel the context when the service call times out
    select{
    case v := <-result:
       action.Log(v, 6)
    case <-action.Done():
        cancel()
    }

    return action, nil
}
```

It is highly recommended to monitor this channel and stop any ongoing task when the channel is closed.

Documentation
-------------

See the [API](http://kusanagi.io/docs/sdk) for a technical reference of the SDK.

For help using the framework see the [documentation](http://kusanagi.io/docs).

Support
-------

Please first read our [contribution guidelines](http://kusanagi.io/open-source/contributing).

* [Requesting help](http://kusanagi.io/open-source/help)
* [Reporting a bug](http://kusanagi.io/open-source/bug)
* [Submitting a patch](http://kusanagi.io/open-source/patch)
* [Security issues](http://kusanagi.io/open-source/security)

We use [milestones](https://github.com/kusanagi/kusanagi-sdk-go/milestones) to track upcoming releases inline with our [versioning](http://kusanagi.io/open-source/roadmap#versioning) strategy, and as defined in our [roadmap](http://kusanagi.io/open-source/roadmap).

Contributing
------------

If you'd like to know how you can help and support our Open Source efforts see the many ways to [get involved](http://kusanagi.io/open-source).

Please also be sure to review our [community guidelines](http://kusanagi.io/open-source/conduct).

License
-------

Copyright 2016-2021 KUSANAGI S.L. (http://kusanagi.io). All rights reserved.

KUSANAGI, the sword logo and the "K" logo are trademarks and/or registered trademarks of KUSANAGI S.L. All other trademarks are property of their respective owners.

Licensed under the [MIT License](https://opensource.org/licenses/MIT). Redistributions of the source code included in this repository must retain the copyright notice found in each file.
