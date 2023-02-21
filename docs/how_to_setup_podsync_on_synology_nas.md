# Podsync on Synology NAS Guide

*Written by [@lucasjanin](https://github.com/lucasjanin)*

This installs `podsync` on a Synology NAS with SSL and port 443
It requires to have a domain with ddns and an SSL Certificate
I'm using a ddns from Synolgy with a SSL Certificate. By chance, my provider doesn't block ports 80 and 443.


1. Open "Package Center" and install "Apache HTTP Server 2.4"
2. In the "Web Station", select the default server, click edit and active "Enable personal website"
3. Create a folder "podsync" in web share using "File Station", the path will be like "/volume1/web/podsync" (where the files will be saved)
4. Create a folder "podsync" in another share using "File Station", the path will be like "/volume1/docker/podsync" (where the config will be saved)
5. Create a `config.toml` file in Notepad (or any other editor) and copy it into the above folder.
Here you will configure your specific settings. Here's mine as an example:

```toml
[server]
port = 9090
hostname = "https://xxxxxxxx.xxx"

[storage]
  [storage.local]
  data_dir = "/app/data" 

[tokens]
youtube = "xxxxxxx"

[feeds]
    [feeds.ID1]
    url = "https://www.youtube.com/channel/UCJldRgT_D7Am-ErRHQZ90uw"
    update_period = "1h"
    quality = "high" # "high" or "low"
    format = "audio" # "audio", "video" or "custom"
    filters = { title = "Yann Marguet" }
    opml = true
    clean = { keep_last = 20 }
    private_feed = true
    [feeds.ID1.custom]
    title = "Yann Marguet - Moi, ce que j'en dis..."
    description = "Yann Marguet sur France Inter"
    author = "Yann Marguet"
    cover_art = "https://www.radiofrance.fr/s3/cruiser-production/2023/01/834dd18e-a74c-4a65-afb0-519a5f7b11c1/1400x1400_moi-ce-que-j-en-dis-marguet.jpg"
    cover_art_quality = "high"
    category = "Comedy"
    subcategories = ["Stand-Up"]
    lang = "fr"
    ownerName = "xxxx xxxxx"
    ownerEmail = "xx@xxxx.xx"
```

Note that I'm not using port `8080` because I already have another app on my Synology using that port.
Also, I'm using my own hostname so I can download the podcasts to my podcast app from outside my network,
but you don't need to do this.

6. Now you need to SSH into Synology using an app like Putty (on Windows - just google for an app).

5. Copy and paste the following command:

```bash
docker pull mxpv/podsync:latest
```

Docker will download the latest version of Podsync.

6. Copy and paste the following command:

```bash
docker run \
    -p 9090:9090 \
    -v /volume1/web/podsync:/app/data/ \
    -v /volume1/docker/podsync/podsync-config.toml:/app/config.toml \
    mxpv/podsync:latest
```

This will install a container in Docker and run it. Podsync will load and read your config.toml file and start downloading episodes.

7. I recommend you go into the container's settings in Container Station and set it to Auto Start.

8. Once the downloads have finished for each of your feeds, you will then have an XML feed for each feed
that you should be able to access at `https://xxxxxxxx.xxx/podsync/ID1.xml`. Paste them into your podcast app of choice,
and you're good to go!

Note: you can validate your XML using this website:
https://www.castfeedvalidator.com/validate.php
