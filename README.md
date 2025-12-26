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

           ✓ Upload complete!

   bURL:  https://api.bucketlabs.org/d/bk9b360f45-f40
 Secret:  4d646d88-6f83-41
Expires:  2026-06-25T00:00:00Z
```

<p align="right">
  <img src="./img/msg.png" />
</p>

and your colleagues:
```sh
$ bucket pull https://api.bucketlabs.org/d/bk9b360f45-f40
Enter secret: 
Downloaded: Rocky-9.7-Custom-KS-x86_64-minimal.iso
```

## Installation
Install in seconds:
```sh
curl -sSL bucketlabs.org/install.sh | bash
```
