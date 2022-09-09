# LunarFuzz

Fuzzer for Dynamic JS application, such as Angular apps, that require a browser that render the JS and thus can not be fuzzed with tools like ffuf (afaik). LunarFuzz uses go-rod, a selenium-like toolkit which runs a headless browser that does the requests.

However, I don't know shit about go, so dont expect good code. Python was too slow, Nim didn't have a good library for a webdriver and I hate writing Rust-code with a passion, so here we are.

Usage examples (subject to change):

```bash
# Autocalibrate, save screenshots on match
lunarfuzz -u https://target.site -w /usr/share/wordlists/dirb/big.txt --screenshot -b "PHPSESSID=XYZ; __OTHER_COOKIE=1"
# filter by sizes
lunarfuzz -u https://target.site -w /usr/share/wordlists/dirb/big.txt --screenshot -b "PHPSESSID=XYZ; __OTHER_COOKIE=1" -fs 1000,1001
```