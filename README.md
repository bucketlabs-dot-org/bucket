# 🪣 bucket
## Share Files Like You Ship Code


<p align="right">
  <img src="./img/bkt-dashboard.png" />
</p>

Say you need to share your new Rocky 9.7 iso with your team. 

```
WeTransfer?   🙄   Too tedious...

Teams?        🤦   Too large...

OneDrive?     🤬   #%!
```

### What if...

you:
```sh
$ bucket push ./isos/Rocky-9.7-Custom-KS-x86_64-minimal.iso
Upload complete!
Tiny URL: https://bucket.io/d/bcd0604ab
Download secret: 4acf38cdd17a5a73
Expires: 2025-12-07T18:17:30.780403029Z
```

<p align="right">
  <img src="./img/msg.png" />
</p>

and your colleagues:
```sh
$ bucket pull https://bucket.io/d/bcd0604ab
Enter secret: 
Downloaded: Rocky-9.7-Custom-KS-x86_64-minimal.iso
```

## Installation
Install in seconds:
```sh
curl -sSL bucketlabs.org/bucket.sh | bash
```
