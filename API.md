## Using bandwidth hero endpoint

When you first request to the backend with no parameter, It should respond like this:
```
bandwidth-hero-proxy
```

### Query Parameter
There's this query parameter that you should know:
- `url` (string): QueryEscaped URL to the image source
- `bw` (uint) (optional): Toggle between transforming color profile to black and white color (default: `1`) (value: `0` for color, `1` for black and white)
- `l` (uint) (optional): Range of image quality between 1-100. (default: `80`)
- `a` (uint) (optional): Toggle whenever to convert GIF to animated webp. (default: `1`) (value: `1` for animated, `0` for static)
- `nr` (uint) (optional): Toggle for no-redirect to the source image when this is viewed in browser's new tab (default: `0`) (value: `1` for no redirect, `0` for redirect)

Example:
```
https://waltuh.cyou/bwhero/
  ?url=https%3A%2F%2Fmedia.fedinet.waltuh.cyou%2Fproxy%2Fpreview%2FQ1oJEV0xzkKd6ZfdIjEjje3Kils%2FaHR0cHM6Ly9tZWRpYS5mb3BzLmNsb3VkL21lZGlhL2RmZmI2NmIwZGQ3MzlhNDg2Yzg2ZmM4ZDczZDg3ZTdlYTY2YjcwMGYzODNkYjEyMDI4ZTcyYmFmNDVmY2MyZDUucG5n%2Fdffb66b0dd739a486c86fc8d73d87e7ea66b700f383db12028e72baf45fcc2d5.png
  &bw=0
  &l=60
  &nr=1
  &a=1
```