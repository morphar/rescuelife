# rescuelife
Rescue as much as possible from your Picturelife account

## Install & run
The binary keeps a file with status of fetches on the disk.  
Because of this, you can restart the application as many times as you like.  

It may cause some downloads to fail, but you should probably retry all failed downloads anyways, after you have successfully run to the end once.

If the process doesn't start immediately, the Picturelife servers are probably down, the application will retry for a while and then die.

You can safely re-run the application.

### If you have Go(lang) installed
Install:  
```go get github.com/morphar/rescuelife```  

Run:  
```rescuelife -help```


### If you are on OS X / macOS
You can download a pre-build binary [here](https://github.com/morphar/rescuelife/releases).  
Or [direct link](https://github.com/morphar/rescuelife/releases/download/0.1.0/rescuelife) to the binary.

In the Finder, double click on the downloaded file.

Alternative:  
Open your terminal, change dir to where you downloaded the binary, then run:  
```./rescuelife -help```

## Notes
It seems like some pictures and videos are totally gone from the service.  
This is why the applications skips already tried and failed images and videos.  
You can retry failed images and videos, by adding the flag: ```-retry```

