# `LAB`

Gitlab command line tool

## INSTALLATION

`$ go get -u github.com/educas/lab`

## USAGE

```
$ lab mr

# NAME:
#    lab merge-request - do something with merge requests
# 
# USAGE:
#    lab merge-request command [command options] [arguments...]
# 
# COMMANDS:
#    list		list merge requests
#    create	create a merge request
#    accept	accept merge request by the current branch
#    help, h	Shows a list of commands or help for one command
#    
# OPTIONS:
#    --help, -h	show help
   
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
