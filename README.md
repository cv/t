# t

Quick and dirty world clock for the command-line.

Install with:

```bash
go install github.com/cv/t@master
```

And then:

```bash
$ t sfo jfk
SFO: 🕓  16:06:21 (America/Los_Angeles)
JFK: 🕖  19:06:21 (America/New_York)
```

Any IATA airport code can be used, and will pick the timezone of that airport.

If `PS1_FORMAT` is set, the output will be simply the airport code and the time, with no decorations or newline:

```bash
$ echo $(PS1_FORMAT=1 t sfo lon)
SFO 17:47 LON 01:47
```

That's it!

## License

Copyright 2017 Carlos Villela <cv@lixo.org>

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
