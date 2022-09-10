# LunarFuzz

Fuzzer for Dynamic JS applications, such as Angular apps, that require a browser that renders the JS and thus can not be fuzzed with tools like ffuf (afaik). LunarFuzz uses go-rod, a selenium-like toolkit which runs a headless browser that does the requests.

### Usage

```
.-.   .-. .-..-. .-.  .--.  .----. .----..-. .-. .---. .---. 
| |   | { } ||  `| | / {} \ | {}  }| {_  | { } |{_   /{_   / 
| `--.| {_} || |\  |/  /\  \| .-. \| |   | {_} | /    }/    }
`----'`-----'`-' `-'`-'  `-'`-' `-'`-'   `-----' `---' `---'
LunarFuzz v0.0.1

usage: lunarfuzz [-h|--help] -u|--url "<value>" -w|--wordlist "<value>" [--fs
                 "<value>"] [--fw "<value>"] [--fm "<value>"] [-b|--cookies
                 "<value>"] [-H|--Headers "<value>"] [--screenshot]
                 [-t|--threads <integer>] [--no-ac] [--2f2f] [-o|--output-file
                 "<value>"]

                 Directory fuzzer for dynamic JS & single page apps

Arguments:

  -h  --help         Print help information
  -u  --url          Target url
  -w  --wordlist     Wordlist to use
      --fs           Filter responses by size. Can also specify multiple, e.g.
                     80,102
      --fw           Filter responses by word count. Can also specify multiple,
                     e.g. 100,101,102
      --fm           Filter responses by substring match. Can also specify
                     multiple, e.g. '404,Not found'
  -b  --cookies      Cookies to use in the format of 'authToken=abcdefg;
                     __otherCookie=1
  -H  --Headers      Headers to use in the format of 'Header: Value; Header:
                     Value'
      --screenshot   Save screenshots for matches. Default: false
  -t  --threads      Number of threads. Default: 5
      --no-ac        Do not run autocalibration if no filter is given. Will
                     output every url as a finding. Default: false
      --2f2f         Do not wait for page to render completely. Default: false
  -o  --output-file  File to save all matching urls to
```

Examples:

```bash
# Autocalibrate, add cookies, save screenshots on match
lunarfuzz -u https://target.site -w /usr/share/wordlists/dirb/big.txt --screenshot -b "SESSION=XYZ; __OTHER_COOKIE=1"
# filter by sizes
lunarfuzz -u https://target.site -w /usr/share/wordlists/dirb/big.txt -fs 1000,1001
# filter by string match, use a custom header, save output to file and use 20 threads
lunarfuzz -u https://target.site -w /usr/share/wordlists/dirb/big.txt -fm "404,not found" -t 20 -H "Authorization: Basic ZGVlejpudXRz"
```


##### Disclaimer 

I don't know shit about go, so dont expect good/performant code. But python was too slow, Nim didn't have a good library for a webdriver and I hate writing Rust-code with a passion, so here we are.

