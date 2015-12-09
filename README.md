# `LAB`

Gitlab command line tool

## INSTALLATION

`$ go get -u github.com/ordbogen/lab`

Export your gitlab private token as environment variable: `LAB_PRIVATE_TOKEN`

```bash
# Get from gitlab.server.com/profile/account
$ export LAB_PRIVATE_TOKEN my-private-token
```

## USAGE

```
$ lab help

# ...
# COMMANDS:
#    browse             Open project homepage
#    merge-request, mr  Merge requests: create, list, browse, checkout, accept, ...
#    help, h            Shows a list of commands or help for one command
# ...


$ lab mr help

# ...
# COMMANDS:
#    create, c     Create merge request, default target branch: master.
#    browse, b     Browse current merge request or by ID.
#    accept        Accept current merge request or by ID.
#    diff          Diff current merge request or by ID.
#    pick-diff     Pick diff from merge requests
#    list, l       List merge requests
#    checkout, co  Checkout branch from merge request
#    help, h       Shows a list of commands or help for one command
# ...
```

## IDEAS

- [x] `$ lab mr browse` -> Open the current merge-request (current branch on the left)
- [x] `$ lab browse` -> open project page
- [x] Show url for private token is missing
- [ ] Fancy rendering/interactivity via [github.com/nsf/termbox-go](https://github.com/nsf/termbox-go)
- [ ] Use goconvey for testing
- [ ] Web interface

## LICENSE
 
```
The MIT License (MIT)

Copyright (c) Thomas B Homburg 2014

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
