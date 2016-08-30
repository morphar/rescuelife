# OBSOLETE!
## After Picturelife shut down, this tool is no longer functional

# rescuelife
Rescue as much as possible from your Picturelife account

## Install & run
The binary keeps a file with status of fetches on the disk.  
Because of this, you can restart the application as many times as you like.  

It may cause some downloads to fail, but you should probably retry all failed downloads anyways, after you have successfully run to the end once.

If the process doesn't start immediately, the Picturelife servers are probably down, the application will retry for a while and then die.

You can safely re-run the application.

### If you are on OS X / macOS
You can download a pre-build binary [here](https://github.com/morphar/rescuelife/releases).  
Or [direct link](https://github.com/morphar/rescuelife/releases/download/0.3.0/rescuelife) to the binary.

In the Finder, double click on the downloaded file.

##### If that doesn't work:
Open Terminal app, change dir to where you downloaded the binary, then run:  
```
cd Downloads        # Takes you to your download dir
chmod +x rescuelife # This will make the file executable
./rescuelife -help  # Help text about what flags can be used
./rescuelife        # This will run the program
```

You can always run the program again to see if any more files has become avialable:
```./rescuelife -retry```

### If you have Go(lang) installed
Install:  
```go get github.com/morphar/rescuelife```  

Run:  
```rescuelife -help```

## Notes
It seems like some pictures and videos are totally gone from the service.  
This is why the applications skips already tried and failed images and videos.  
You can retry failed images and videos, by adding the flag: ```-retry```  

My personal final result was 4656 out of 9504 files.
