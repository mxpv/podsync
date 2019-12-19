# Podsync on QNAP NAS Guide

*Written by [@Rumik](https://github.com/Rumik)*

1. Install Container Station from App Center.
2. Create a shared folder on your QNAP for where you want Podsync to store its config file and data,
e.g. `/share/CACHEDEV1_DATA/appdata/podsync`
3. Create a `config.toml` file in Notepad or whatever editor you want to use and copy it into the above folder.
Here you will configure your specific settings. Here's mine as an example:

```toml
[server]
port = 6969
data_dir = "/share/CACHEDEV1_DATA/appdata/podsync"
hostname = "http://my.customhostname.com:6969"

[tokens]
youtube = "INSERTYOUTUBEAPI" # Tokens from `Access tokens` section

[feeds]
  [feeds.KFGD] # Kinda Funny Games Daily
  url = "youtube.com/playlist?list=PLy3mMHt2i7RIl9pkdvrA98kN-RD4yoRhv"
  page_size = 3
  update_period = "60m"
  quality = "high"
  format = "video"
  cover_art = "http://i1.sndcdn.com/avatars-000319281278-0merek-original.jpg"
```

Note that I'm not using port `8080` because I already have another app on my QNAP using that port.
I'm using port `6969` specifically because `Bill & Ted!`.
Also, I'm using my own hostname so I can download the podcasts to my podcast app from outside my network,
but you don't need to do this. To make that work, make sure you forward port `6969` to your QNAP.

4. By now, Container Station should have finished installing and should now be running.
Now you need to SSH into the QNAP using an app like Putty (on Windows - just google for an app).

5. Copy and paste the following command:

```bash
docker pull mxpv/podsync:latest
```

Docker will download the latest version of Podsync.

6. Copy and paste the following command:

```bash
docker run \
    -p 6969:6969 \
    -v /share/CACHEDEV1_DATA/appdata/podsync:/app/data/ \
    -v /share/CACHEDEV1_DATA/appdata/podsync/config.toml:/app/config.toml \
    mxpv/podsync:latest
```

This will install a container in Container Station and run it. Podsync will load and read your config.toml file and start downloading episodes.

7. I recommend you go into the container's settings in Container Station and set it to Auto Start.

8. Once the downloads have finished for each of your feeds, you will then have an XML feed for each feed
that you should be able to access at `http://ipaddressorhostname:6969/`. Paste them into your podcast app of choice,
and you're good to go!
