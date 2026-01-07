# 🪣 bucket
## Share Files Like You Ship Code

Say you need to share your new 16GB Rocky 9.7 iso with your systems team.  

```
WeTransfer?   🙄   Too tedious...

Teams?        🤦   Too large...

OneDrive?     🤬   #%!
```

### What if you just had to run a simple command...
```sh
$ bucket push ./isos/Rocky-9.7-Custom-KS-x86_64-minimal.iso

           ✓ Upload complete!
    bID:  bk9b360f45-f40
   bURL:  https://api.bucketlabs.org/d/bk9b360f45-f40
 Secret:  4d646d88-6f83-41
Expires:  2026-06-25T00:00:00Z
```

<p align="center">
  <img src="./img/msg.png" />
</p>

### and your technical colleagues could do the same:
```sh
$ bucket pull bk9b360f45-f40
Enter secret: 
Downloaded: Rocky-9.7-Custom-KS-x86_64-minimal.iso
```

### your non-technical colleagues can just click the bURL link and enter the secret key:
<p align="center">
  <img src="./img/downloads.png" />
</p>


### Not a terminal person? We have a dashboard for you.
<p align="right">
  <img src="./img/bkt-dashboard.png" />
</p>


## Installation
Install in seconds:
```sh
# WSL/MacOS/Linux
curl -sSL bucketlabs.org/install.sh | bash
```

```powershell
# Windows (powershell)
irm bucketlabs.org/install.ps1 | iex 
```
